package emulator

import (
	"context"
	"fmt"
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
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/transport/tcp"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/util"
)

type Service struct {
	cfg    *config.EmulatorConfig
	mode   checksum.Mode
	wire   codec.Codec
	logger *log.Logger
	fault  FaultMode
	now    func() time.Time

	history *events.Ring
	server  *tcp.Server

	mu     sync.RWMutex
	status Status
}

func NewService(cfg *config.EmulatorConfig, logger *log.Logger) (*Service, error) {
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
		fault:   FaultModeFromConfig(cfg),
		now:     time.Now,
		history: events.NewRing(cfg.RecentSize),
	}
	s.status = Status{
		ListenAddress: cfg.ListenAddress(),
		ChecksumMode:  string(mode),
		FaultMode:     s.fault,
	}
	return s, nil
}

func (s *Service) Run(ctx context.Context) error {
	s.setRunning(true)
	defer s.setRunning(false)

	server := tcp.NewServer(tcp.ServerConfig{Address: s.cfg.ListenAddress()}, s, s.logger)
	s.mu.Lock()
	s.server = server
	s.mu.Unlock()
	s.logger.Printf("emulator service listening on %s mode=%s", s.cfg.ListenAddress(), s.mode)
	return server.Run(ctx)
}

func (s *Service) Addr() net.Addr {
	s.mu.RLock()
	server := s.server
	s.mu.RUnlock()
	if server == nil {
		return nil
	}
	return server.Addr()
}

func (s *Service) HandleConnection(ctx context.Context, conn net.Conn) {
	s.connectionOpened()
	defer s.connectionClosed()

	session := newSession(s, conn)
	session.run(ctx)
}

func (s *Service) BuildResponse(req frame.Frame) ([]byte, string, bool, error) {
	data := req.DataBytes()
	if err := command.ParseReadTimeRequest(data); err == nil {
		resp, err := s.wire.EncodeReadTimeResponse(req, s.now())
		return resp, "read-time", true, err
	}

	resp, err := s.wire.EncodeACK(req, data)
	cmd := "ack"
	if len(data) > 0 {
		cmd = fmt.Sprintf("unknown-0x%02X", data[0])
	}
	return resp, cmd, false, err
}

func (s *Service) Mode() checksum.Mode {
	return s.mode
}

func (s *Service) FaultMode() FaultMode {
	return s.fault
}

func (s *Service) Timeouts() (time.Duration, time.Duration) {
	return s.cfg.ReadTimeoutDuration, s.cfg.WriteTimeoutDuration
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

func (s *Service) recordFrame(direction, remote, rawHex, cmd, errText string) {
	s.history.Add(events.FrameRecord{
		Timestamp:    s.now(),
		Direction:    direction,
		Service:      "emulator",
		RemoteAddr:   remote,
		RawHex:       rawHex,
		Command:      cmd,
		ChecksumMode: string(s.mode),
		Error:        errText,
	})
}

func (s *Service) recordRX(remote string, req frame.Frame) {
	s.recordFrame("RX", remote, util.HexDump(req.RawBytes()), commandName(req.DataBytes()), "")
	s.mu.Lock()
	s.status.TotalRequests++
	s.status.LastRequestTime = s.now()
	s.mu.Unlock()
}

func (s *Service) recordTX(remote string, raw []byte, cmd string) {
	s.recordFrame("TX", remote, util.HexDump(raw), cmd, "")
	s.mu.Lock()
	s.status.TotalResponses++
	s.status.LastResponseTime = s.now()
	s.mu.Unlock()
}

func (s *Service) recordError(remote string, err error) {
	errText := ""
	if err != nil {
		errText = err.Error()
	}
	s.recordFrame("ERR", remote, "", "", errText)
	s.mu.Lock()
	s.status.TotalProtocolErrors++
	s.status.LastError = errText
	s.mu.Unlock()
}

func (s *Service) connectionOpened() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.ActiveConnections++
	s.status.TotalConnections++
}

func (s *Service) connectionClosed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status.ActiveConnections > 0 {
		s.status.ActiveConnections--
	}
}

func (s *Service) setRunning(running bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.Running = running
}

func commandName(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	if data[0] == byte(command.ReadTime) {
		return "read-time"
	}
	return fmt.Sprintf("unknown-0x%02X", data[0])
}
