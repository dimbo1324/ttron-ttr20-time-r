package gateway

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/clock"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/command"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/schedule"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/util"
)

const readBufferSize = 4096

type pollSession struct {
	service *Service
	conn    net.Conn
	parser  *frame.StreamParser
	remote  string
	buffer  []byte
}

func (s *Service) newPollSession(conn net.Conn) *pollSession {
	return &pollSession{
		service: s,
		conn:    conn,
		parser:  frame.NewStreamParser(s.mode),
		remote:  conn.RemoteAddr().String(),
		buffer:  make([]byte, readBufferSize),
	}
}

func (s *Service) runSession(ctx context.Context, conn net.Conn) error {
	session := s.newPollSession(conn)
	if s.cfg.IdentityProbe {
		session.probeIdentity()
	}
	if s.schedule.Mode() != schedule.ModeAligned {
		if err := session.pollWithRetry(ctx); err != nil {
			return err
		}
	}

	ticker := schedule.NewTicker(s.schedule)
	for {
		next := ticker.NextAt()
		s.nextPollAt(next)
		if _, ok := ticker.WaitUntil(ctx, next); !ok {
			return nil
		}
		if err := session.pollWithRetry(ctx); err != nil {
			return err
		}
	}
}

func (p *pollSession) pollWithRetry(ctx context.Context) error {
	service := p.service
	var last error

	for attempt := 0; attempt <= service.retry.Attempts; attempt++ {
		if attempt > 0 {
			delay := service.retry.DelayFor(attempt)
			service.incrementRetries()
			service.logger.Printf("gateway retry attempt=%d/%d delay=%s cause=%v",
				attempt, service.retry.Attempts, delay, last)
			if !sleepContext(ctx, delay) {
				return nil
			}
		}

		err := p.pollOnce()
		if err == nil {
			return nil
		}
		last = err
		if !retryableOnSameConnection(err) {
			service.recordFailure(err)
			service.observeHealthFailure(err)
			return err
		}
	}

	service.incrementExhaustedPolls()
	failure := fmt.Errorf("%w after %d attempts: %v", ErrPollExhausted, service.retry.Attempts+1, last)
	service.recordFailure(failure)
	service.observeHealthFailure(failure)
	service.logger.Printf("gateway poll exhausted target=%s last=%v", service.cfg.Target, last)
	return nil
}

func (p *pollSession) pollOnce() error {
	service := p.service

	request, err := service.wire.EncodeReadTimeRequest()
	if err != nil {
		return err
	}

	deadline := time.Now().Add(service.cfg.RequestTimeout)
	if err := p.conn.SetWriteDeadline(deadline); err != nil {
		return fmt.Errorf("set write deadline: %w", err)
	}
	sentAt := time.Now()
	if _, err := p.conn.Write(request); err != nil {
		return fmt.Errorf("write request: %w", err)
	}
	service.recordTX(p.remote, request, command.NameReadTime)
	service.logger.Printf("gateway TX: %s", util.HexDump(request))

	raw, receivedAt, err := p.readFrame(deadline)
	if err != nil {
		return err
	}
	service.recordRX(p.remote, raw, command.NameReadTime)
	service.logger.Printf("gateway RX: %s", util.HexDump(raw))

	_, parsed, err := service.wire.DecodeReadTimeResponse(raw)
	if err != nil {
		service.incrementProtocolErrors()
		service.recordProtocolError(p.remote, err)
		return fmt.Errorf("decode read-time response: %w", err)
	}

	roundTrip := receivedAt.Sub(sentAt)
	service.recordRoundTrip(roundTrip)
	service.recordSuccess(parsed)
	service.observeHealthSuccess(roundTrip)
	service.observeClock(clock.Sample{RequestedAt: sentAt, ReceivedAt: receivedAt, DeviceTime: parsed.Time})
	service.logger.Printf("gateway device time: %s rtt=%s", parsed.Time.Format(time.RFC3339), roundTrip)
	return nil
}

func (p *pollSession) readFrame(deadline time.Time) ([]byte, time.Time, error) {
	service := p.service
	for {
		if err := p.conn.SetReadDeadline(deadline); err != nil {
			return nil, time.Time{}, fmt.Errorf("set read deadline: %w", err)
		}
		count, readErr := p.conn.Read(p.buffer)
		if count > 0 {
			receivedAt := time.Now()
			result := p.parser.Push(p.buffer[:count])
			for _, parseErr := range result.Errors {
				service.incrementProtocolErrors()
				service.recordProtocolError(p.remote, parseErr)
				service.logger.Printf("gateway protocol parse error: %v", parseErr)
			}
			if len(result.Frames) > 0 {
				return result.Frames[0].RawBytes(), receivedAt, nil
			}
			if len(result.Errors) > 0 {
				return nil, time.Time{}, result.Errors[0]
			}
		}
		if readErr != nil {
			if isTimeoutError(readErr) {
				return nil, time.Time{}, fmt.Errorf("%w: %v", ErrNoResponse, readErr)
			}
			return nil, time.Time{}, fmt.Errorf("read response: %w", readErr)
		}
	}
}

func (p *pollSession) probeIdentity() {
	service := p.service
	if service.identityKnown() || !service.identitySupported() {
		return
	}
	if !service.commands.Supports(command.ReadIdentity) {
		return
	}

	request, err := service.wire.EncodeReadIdentityRequest()
	if err != nil {
		service.logger.Printf("gateway identity probe skipped: %v", err)
		return
	}

	deadline := time.Now().Add(service.cfg.RequestTimeout)
	if err := p.conn.SetWriteDeadline(deadline); err != nil {
		service.logger.Printf("gateway identity probe failed: %v", err)
		return
	}
	if _, err := p.conn.Write(request); err != nil {
		service.logger.Printf("gateway identity probe failed: %v", err)
		return
	}
	service.recordTX(p.remote, request, command.NameReadIdentity)

	raw, receivedAt, err := p.readFrame(deadline)
	if err != nil {
		service.logger.Printf("gateway identity probe unanswered: %v", err)
		return
	}
	service.recordRX(p.remote, raw, command.NameReadIdentity)

	_, identity, err := service.wire.DecodeReadIdentityResponse(raw)
	if err != nil {
		service.logger.Printf("gateway identity not supported by device: %v", err)
		service.markIdentityUnsupported()
		return
	}

	service.recordIdentity(identity, receivedAt)
	service.recordIdentityEvent(identity, receivedAt)
	service.logger.Printf("gateway device identity model=%s serial=%s firmware=%s",
		identity.Model, identity.Serial, identity.Firmware)
}
