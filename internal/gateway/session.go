package gateway

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/util"
)

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
	s.logger.Printf("gateway polling request command=read-time target=%s timeout=%s", conn.RemoteAddr().String(), s.cfg.RequestTimeout)
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
			s.logger.Printf("gateway read timeout/error: %v", err)
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
