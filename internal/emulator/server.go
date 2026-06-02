package emulator

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"runtime/debug"
	"sync"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/codec"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/command"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/util"
)

type Server struct {
	cfg    *config.EmulatorConfig
	logger *log.Logger
	ln     net.Listener
	wg     sync.WaitGroup
	close  chan struct{}
	closed bool
	mu     sync.Mutex
}

func NewServer(cfg *config.EmulatorConfig, logger *log.Logger) *Server {
	return &Server{cfg: cfg, logger: logger, close: make(chan struct{})}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.ln = ln
	s.logger.Printf("listening on %s", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-s.close:
				return nil
			default:
				s.logger.Printf("accept error: %v", err)
				continue
			}
		}
		s.logger.Printf("accepted connection from %s", conn.RemoteAddr())
		s.wg.Add(1)
		go func(c net.Conn) {
			defer s.wg.Done()
			s.handleConnection(c)
		}(conn)
	}
}

func (s *Server) Stop() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	s.mu.Unlock()

	close(s.close)
	if s.ln != nil {
		_ = s.ln.Close()
	}
	s.logger.Printf("closing server, waiting for handlers...")
	s.wg.Wait()
	s.logger.Printf("server stopped")
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Printf("[%s] PANIC recovered: %v\n%s", conn.RemoteAddr(), r, string(debug.Stack()))
			_ = conn.Close()
		}
	}()
	defer func() {
		_ = conn.Close()
		s.logger.Printf("[%s] connection handler finished", conn.RemoteAddr())
	}()

	mode, err := checksum.ParseMode(s.cfg.CRCMode)
	if err != nil {
		s.logger.Printf("[%s] invalid checksum mode: %v", conn.RemoteAddr(), err)
		return
	}
	parser := frame.NewStreamParser(mode)
	wire := codec.New(mode, 0x00, byte(s.cfg.AdapterAddr&0xFF))
	tmp := make([]byte, 4096)
	readTimeout := time.Duration(s.cfg.ReadTimeout) * time.Second

	for {
		_ = conn.SetReadDeadline(time.Now().Add(readTimeout))
		n, err := conn.Read(tmp)
		if err != nil {
			s.logger.Printf("[%s] read error: %v", conn.RemoteAddr(), err)
			return
		}
		if n == 0 {
			continue
		}

		result := parser.Push(tmp[:n])
		for _, parseErr := range result.Errors {
			s.logger.Printf("[%s] protocol parse error: %v", conn.RemoteAddr(), parseErr)
		}

		for _, req := range result.Frames {
			s.logger.Printf("[%s] RX: %s", conn.RemoteAddr(), util.HexDump(req.RawBytes()))
			resp, ok := s.buildResponse(conn, wire, req)
			if !ok {
				continue
			}

			if s.cfg.DelayMs > 0 {
				time.Sleep(time.Duration(s.cfg.DelayMs) * time.Millisecond)
			}
			if rand.Float64() < s.cfg.BadCRCProb {
				s.logger.Printf("[%s] injecting bad CRC", conn.RemoteAddr())
				frame.CorruptChecksum(resp, string(mode))
			}

			if rand.Float64() < s.cfg.FragProb && len(resp) > 1 {
				if err := writeFragmented(conn, resp); err != nil {
					s.logger.Printf("[%s] write error: %v", conn.RemoteAddr(), err)
					return
				}
			} else if _, err := conn.Write(resp); err != nil {
				s.logger.Printf("[%s] write error: %v", conn.RemoteAddr(), err)
				return
			}
			s.logger.Printf("[%s] TX: %s", conn.RemoteAddr(), util.HexDump(resp))
		}
	}
}

func (s *Server) buildResponse(conn net.Conn, wire codec.Codec, req frame.Frame) ([]byte, bool) {
	if err := command.ParseReadTimeRequest(req.DataBytes()); err == nil {
		s.logger.Printf("[%s] read-time request (ctrl=0x%02X addr=0x%02X)", conn.RemoteAddr(), req.Control, req.Address)
		resp, err := wire.EncodeReadTimeResponse(req, time.Now())
		if err != nil {
			s.logger.Printf("[%s] read-time response encode failed: %v", conn.RemoteAddr(), err)
			return nil, false
		}
		return resp, true
	}

	data := req.DataBytes()
	cmd := byte(0x00)
	if len(data) > 0 {
		cmd = data[0]
	}
	s.logger.Printf("[%s] generic/unknown cmd 0x%02X - sending ACK", conn.RemoteAddr(), cmd)
	resp, err := wire.EncodeACK(req, data)
	if err != nil {
		s.logger.Printf("[%s] ACK encode failed: %v", conn.RemoteAddr(), err)
		return nil, false
	}
	return resp, true
}

func writeFragmented(conn net.Conn, data []byte) error {
	i := len(data) / 2
	if i < 1 {
		i = 1
	}
	if _, err := conn.Write(data[:i]); err != nil {
		return err
	}
	time.Sleep(40 * time.Millisecond)
	_, err := conn.Write(data[i:])
	return err
}
