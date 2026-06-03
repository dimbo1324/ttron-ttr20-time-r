package server

import (
	"context"
	"errors"
	"net"
	"sync"

	"google.golang.org/grpc"
)

type RegisterFunc func(*grpc.Server)

type Server struct {
	address  string
	register RegisterFunc

	mu     sync.RWMutex
	ln     net.Listener
	server *grpc.Server
}

func New(address string, register RegisterFunc, opts ...grpc.ServerOption) *Server {
	s := grpc.NewServer(opts...)
	return &Server{address: address, register: register, server: s}
}

func (s *Server) Run(ctx context.Context) error {
	if s.register != nil {
		s.register(s.server)
	}
	ln, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.ln = ln
	s.mu.Unlock()

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			s.server.GracefulStop()
		case <-done:
		}
	}()

	err = s.server.Serve(ln)
	close(done)
	if ctx.Err() != nil || errors.Is(err, grpc.ErrServerStopped) {
		return nil
	}
	return err
}

func (s *Server) Addr() net.Addr {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ln == nil {
		return nil
	}
	return s.ln.Addr()
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}
