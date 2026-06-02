package emulator

import (
	"context"
	"log"
	"net"
	"sync"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
)

type Server struct {
	service *Service
	cancel  context.CancelFunc
	mu      sync.Mutex
}

func NewServer(cfg *config.EmulatorConfig, logger *log.Logger) *Server {
	service, err := NewService(cfg, logger)
	if err != nil {
		if logger != nil {
			logger.Printf("emulator service config error: %v", err)
		}
		cfg.CRCMode = "sum"
		service, _ = NewService(cfg, logger)
	}
	return &Server{service: service}
}

func (s *Server) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	s.cancel = cancel
	s.mu.Unlock()
	return s.service.Run(ctx)
}

func (s *Server) Stop() {
	s.mu.Lock()
	cancel := s.cancel
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (s *Server) Service() *Service {
	return s.service
}

func (s *Server) Addr() net.Addr {
	if s.service == nil {
		return nil
	}
	return s.service.Addr()
}
