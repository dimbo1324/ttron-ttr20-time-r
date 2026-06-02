package tcp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

type Handler interface {
	HandleConnection(ctx context.Context, conn net.Conn)
}

type HandlerFunc func(ctx context.Context, conn net.Conn)

func (f HandlerFunc) HandleConnection(ctx context.Context, conn net.Conn) {
	f(ctx, conn)
}

type ServerConfig struct {
	Address string
}

type Server struct {
	cfg     ServerConfig
	handler Handler
	logger  *log.Logger

	mu sync.RWMutex
	ln net.Listener
}

func NewServer(cfg ServerConfig, handler Handler, logger *log.Logger) *Server {
	return &Server{cfg: cfg, handler: handler, logger: logger}
}

func (s *Server) Run(ctx context.Context) error {
	if s.handler == nil {
		return fmt.Errorf("tcp server handler is nil")
	}
	ln, err := net.Listen("tcp", s.cfg.Address)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.ln = ln
	s.mu.Unlock()

	var wg sync.WaitGroup
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = ln.Close()
		case <-done:
		}
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			close(done)
			wg.Wait()
			s.mu.Lock()
			if s.ln == ln {
				s.ln = nil
			}
			s.mu.Unlock()
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
		if s.logger != nil {
			s.logger.Printf("tcp accepted connection from %s", conn.RemoteAddr())
		}
		wg.Add(1)
		go func(c net.Conn) {
			defer wg.Done()
			defer c.Close()
			s.handler.HandleConnection(ctx, c)
		}(conn)
	}
}

func (s *Server) Addr() net.Addr {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ln == nil {
		return nil
	}
	return s.ln.Addr()
}

func (s *Server) Close() error {
	s.mu.RLock()
	ln := s.ln
	s.mu.RUnlock()
	if ln == nil {
		return nil
	}
	return ln.Close()
}
