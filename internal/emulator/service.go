package emulator

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/observability/events"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/codec"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/transport/tcp"
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
	remote := conn.RemoteAddr().String()
	s.logger.Printf("[%s] connection accepted", remote)
	defer func() {
		s.connectionClosed()
		s.logger.Printf("[%s] connection closed", remote)
	}()

	session := newSession(s, conn)
	session.run(ctx)
}

func (s *Service) Mode() checksum.Mode {
	return s.mode
}

func (s *Service) FaultMode() FaultMode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.fault
}

func (s *Service) SetFaultMode(fault FaultMode) FaultMode {
	fault.BadChecksumProb = clampProbability(fault.BadChecksumProb)
	fault.FragmentProb = clampProbability(fault.FragmentProb)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fault = fault
	s.status.FaultMode = fault
	s.logger.Printf("emulator fault mode updated responseDelay=%s badChecksum=%.3f fragment=%.3f fragmentDelay=%s noResponse=%t closeAfterRequest=%t",
		fault.ResponseDelay, fault.BadChecksumProb, fault.FragmentProb, fault.FragmentDelay, fault.NoResponse, fault.CloseAfterRequest)
	return fault
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
