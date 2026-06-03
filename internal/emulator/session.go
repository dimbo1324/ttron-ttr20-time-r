package emulator

import (
	"context"
	"errors"
	"io"
	"net"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/observability/events"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/util"
)

type session struct {
	service *Service
	conn    net.Conn
	remote  string
	parser  *frame.StreamParser
}

func newSession(service *Service, conn net.Conn) *session {
	return &session{
		service: service,
		conn:    conn,
		remote:  conn.RemoteAddr().String(),
		parser:  frame.NewStreamParser(service.Mode()),
	}
}

func (s *session) run(ctx context.Context) {
	readTimeout, _ := s.service.Timeouts()
	buf := make([]byte, 4096)

	for {
		if ctx.Err() != nil {
			return
		}
		if readTimeout > 0 {
			_ = s.conn.SetReadDeadline(time.Now().Add(readTimeout))
		}
		n, err := s.conn.Read(buf)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				s.service.recordError(s.remote, err)
				s.service.logger.Printf("[%s] read error: %v", s.remote, err)
			}
			return
		}
		if n == 0 {
			continue
		}

		result := s.parser.Push(buf[:n])
		for _, parseErr := range result.Errors {
			s.service.recordError(s.remote, parseErr)
			s.service.logger.Printf("[%s] protocol parse error: %v", s.remote, parseErr)
		}
		for _, req := range result.Frames {
			s.handleFrame(req)
			if s.service.FaultMode().CloseAfterRequest {
				return
			}
		}
	}
}

func (s *session) handleFrame(req frame.Frame) {
	s.service.recordRX(s.remote, req)
	s.service.logger.Printf("[%s] RX: %s", s.remote, util.HexDump(req.RawBytes()))

	resp, cmd, _, err := s.service.BuildResponse(req)
	if err != nil {
		s.service.recordError(s.remote, err)
		s.service.logger.Printf("[%s] response encode failed: %v", s.remote, err)
		return
	}

	fault := s.service.FaultMode()
	if fault.NoResponse {
		s.service.logger.Printf("[%s] fault no-response applied", s.remote)
		s.service.recordFrame(events.DirectionDrop, s.remote, util.HexDump(resp), cmd, "no response fault")
		return
	}
	if fault.ResponseDelay > 0 {
		s.service.logger.Printf("[%s] fault response delay %s", s.remote, fault.ResponseDelay)
		time.Sleep(fault.ResponseDelay)
	}
	if fault.ShouldCorruptChecksum() {
		s.service.logger.Printf("[%s] fault corrupt checksum applied", s.remote)
		fault.ApplyCorruptChecksum(resp, string(s.service.Mode()))
	}

	if err := s.writeResponse(resp, fault); err != nil {
		s.service.recordError(s.remote, err)
		s.service.logger.Printf("[%s] write error: %v", s.remote, err)
		return
	}
	s.service.recordTX(s.remote, resp, cmd)
	s.service.logger.Printf("[%s] TX: %s", s.remote, util.HexDump(resp))
}

func (s *session) writeResponse(resp []byte, fault FaultMode) error {
	_, writeTimeout := s.service.Timeouts()
	if writeTimeout > 0 {
		_ = s.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	}
	if fault.ShouldFragment() && len(resp) > 1 {
		i := len(resp) / 2
		if i < 1 {
			i = 1
		}
		if _, err := s.conn.Write(resp[:i]); err != nil {
			return err
		}
		if fault.FragmentDelay > 0 {
			time.Sleep(fault.FragmentDelay)
		}
		_, err := s.conn.Write(resp[i:])
		return err
	}
	_, err := s.conn.Write(resp)
	return err
}
