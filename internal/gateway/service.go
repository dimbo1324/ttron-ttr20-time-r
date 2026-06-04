package gateway

import (
	"context"
	"log"
	"sync"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/observability/events"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/codec"
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

func (s *Service) Start(ctx context.Context) {
	s.runMu.Lock()
	defer s.runMu.Unlock()
	if s.cancel != nil {
		s.logger.Printf("gateway polling already running")
		return
	}
	s.logger.Printf("gateway polling start")
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
		s.logger.Printf("gateway polling stop requested while not running")
		s.setRunning(false)
		s.setConnected(false)
		return nil
	}
	s.logger.Printf("gateway polling stop")
	cancel()
	err := <-done
	if err != nil {
		return err
	}
	return nil
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
