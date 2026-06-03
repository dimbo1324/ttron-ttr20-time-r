package gateway

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/observability/events"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/codec"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/command"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
	transporttcp "github.com/dimbo1324/ttron-ttr20-time-r/internal/transport/tcp"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/util"
)

type Service struct {
	cfg    *config.GatewayConfig
	mode   checksum.Mode
	wire   codec.Codec
	logger *log.Logger

	history *events.Ring

	mu     sync.RWMutex
	status Status

	runMu  sync.Mutex
	cancel context.CancelFunc
	done   chan error
}

func NewService(cfg *config.GatewayConfig, logger *log.Logger) (*Service, error) {
	mode, err := checksum.ParseMode(cfg.CRCMode)
	if err != nil {
		return nil, err
	}
	if cfg.RecentSize <= 0 {
		cfg.RecentSize = 100
	}
	s := &Service{
		cfg:     cfg,
		mode:    mode,
		wire:    codec.New(mode, 0x00, byte(cfg.AdapterAddr&0xFF)),
		logger:  logger,
		history: events.NewRing(cfg.RecentSize),
	}
	s.status = Status{
		TargetAddress:   cfg.Target,
		ChecksumMode:    string(mode),
		PollingInterval: cfg.PollInterval,
		RequestTimeout:  cfg.RequestTimeout,
		ConnectTimeout:  cfg.ConnectTimeout,
	}
	return s, nil
}

func (s *Service) Run(ctx context.Context) error {
	s.setRunning(true)
	defer s.setRunning(false)

	backoff := NewBackoff(s.cfg.BackoffInitial, s.cfg.BackoffMax)
	for {
		if ctx.Err() != nil {
			return nil
		}
		s.incrementConnectionAttempts()
		s.logger.Printf("gateway connecting target=%s mode=%s", s.cfg.Target, s.mode)
		conn, err := transporttcp.Dial(ctx, transporttcp.ClientConfig{
			Address:        s.cfg.Target,
			ConnectTimeout: s.cfg.ConnectTimeout,
		})
		if err != nil {
			s.recordFailure(fmt.Errorf("connect failed: %w", err))
			delay := backoff.Next()
			if !sleepContext(ctx, delay) {
				return nil
			}
			continue
		}

		backoff.Reset()
		s.setConnected(true)
		s.logger.Printf("gateway connected target=%s", s.cfg.Target)
		err = s.runSession(ctx, conn)
		_ = conn.Close()
		s.setConnected(false)
		if ctx.Err() != nil {
			return nil
		}
		if err != nil {
			s.recordFailure(err)
			s.incrementReconnects()
			s.logger.Printf("gateway session error: %v", err)
		}
		delay := backoff.Next()
		if !sleepContext(ctx, delay) {
			return nil
		}
	}
}

func (s *Service) Start(ctx context.Context) {
	s.runMu.Lock()
	defer s.runMu.Unlock()
	if s.cancel != nil {
		return
	}
	runCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	s.cancel = cancel
	s.done = done
	go func() {
		done <- s.Run(runCtx)
		s.runMu.Lock()
		if s.done == done {
			s.cancel = nil
			s.done = nil
		}
		s.runMu.Unlock()
	}()
}

func (s *Service) Stop() error {
	s.runMu.Lock()
	cancel := s.cancel
	done := s.done
	s.runMu.Unlock()
	if cancel == nil || done == nil {
		s.setRunning(false)
		s.setConnected(false)
		return nil
	}
	cancel()
	err := <-done
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) runSession(ctx context.Context, conn net.Conn) error {
	ticker := time.NewTicker(s.cfg.PollInterval)
	defer ticker.Stop()

	if err := s.pollOnce(conn); err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.pollOnce(conn); err != nil {
				return err
			}
		}
	}
}

func (s *Service) pollOnce(conn net.Conn) error {
	req, err := s.wire.EncodeReadTimeRequest()
	if err != nil {
		return err
	}
	deadline := time.Now().Add(s.cfg.RequestTimeout)
	_ = conn.SetWriteDeadline(deadline)
	if _, err := conn.Write(req); err != nil {
		return fmt.Errorf("write request: %w", err)
	}
	s.recordTX(conn.RemoteAddr().String(), req, "read-time")
	s.logger.Printf("gateway TX: %s", util.HexDump(req))

	parser := frame.NewStreamParser(s.mode)
	buf := make([]byte, 4096)
	for {
		_ = conn.SetReadDeadline(deadline)
		n, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return err
			}
			return fmt.Errorf("read response: %w", err)
		}
		result := parser.Push(buf[:n])
		for _, parseErr := range result.Errors {
			s.recordProtocolError(conn.RemoteAddr().String(), parseErr)
			s.logger.Printf("gateway protocol parse error: %v", parseErr)
		}
		if len(result.Frames) == 0 {
			continue
		}
		raw := result.Frames[0].RawBytes()
		s.recordRX(conn.RemoteAddr().String(), raw, "read-time")
		s.logger.Printf("gateway RX: %s", util.HexDump(raw))
		_, parsed, err := s.wire.DecodeReadTimeResponse(raw)
		if err != nil {
			return fmt.Errorf("decode read-time response: %w", err)
		}
		s.recordSuccess(parsed)
		s.logger.Printf("gateway device time: %s", parsed.Time.Format(time.RFC3339))
		return nil
	}
}

func (s *Service) Status() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := s.status
	out.RecentFramesCount = s.history.Len()
	return out
}

func (s *Service) Snapshot() Snapshot {
	return Snapshot{Status: s.Status(), Recent: s.history.Snapshot()}
}

func (s *Service) recordTX(remote string, raw []byte, cmd string) {
	now := time.Now()
	s.history.Add(events.FrameRecord{Timestamp: now, Direction: "TX", Service: "gateway", RemoteAddr: remote, RawHex: util.HexDump(raw), Command: cmd, ChecksumMode: string(s.mode)})
	s.mu.Lock()
	s.status.LastTXTimestamp = now
	s.mu.Unlock()
}

func (s *Service) recordRX(remote string, raw []byte, cmd string) {
	now := time.Now()
	s.history.Add(events.FrameRecord{Timestamp: now, Direction: "RX", Service: "gateway", RemoteAddr: remote, RawHex: util.HexDump(raw), Command: cmd, ChecksumMode: string(s.mode)})
	s.mu.Lock()
	s.status.LastRXTimestamp = now
	s.mu.Unlock()
}

func (s *Service) recordProtocolError(remote string, err error) {
	s.history.Add(events.FrameRecord{Timestamp: time.Now(), Direction: "ERR", Service: "gateway", RemoteAddr: remote, ChecksumMode: string(s.mode), Error: err.Error()})
}

func (s *Service) recordSuccess(parsed command.ReadTimeResponse) {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.SuccessfulReads++
	s.status.LastSuccessfulReadTime = now
	s.status.LastParsedDeviceTime = parsed.Time
	s.status.LastError = ""
}

func (s *Service) recordFailure(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.FailedReads++
	if err != nil {
		s.status.LastError = err.Error()
	}
}

func (s *Service) incrementConnectionAttempts() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.ConnectionAttempts++
}

func (s *Service) incrementReconnects() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.ReconnectCount++
}

func (s *Service) setConnected(connected bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.Connected = connected
}

func (s *Service) setRunning(running bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.Running = running
}

func sleepContext(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		return ctx.Err() == nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
