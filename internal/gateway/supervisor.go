package gateway

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/clock"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/devices"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/health"
)

type FleetStatus struct {
	Devices        int
	Running        int
	Online         int
	Degraded       int
	Offline        int
	Unknown        int
	ClockOK        int
	ClockWarn      int
	ClockCritical  int
	ClockUnknown   int
	WorstClockSkew time.Duration
	WorstClockID   string
	Statuses       []Status
}

type worker struct {
	device  devices.Device
	service *Service
}

type Supervisor struct {
	base     config.GatewayConfig
	registry *devices.Registry
	logger   *log.Logger

	mu      sync.RWMutex
	workers map[string]*worker
}

func NewSupervisor(base *config.GatewayConfig, registry *devices.Registry, logger *log.Logger) (*Supervisor, error) {
	if base == nil {
		return nil, fmt.Errorf("supervisor requires a base gateway config")
	}
	if registry == nil {
		return nil, fmt.Errorf("supervisor requires a device registry")
	}
	supervisor := &Supervisor{
		base:     *base,
		registry: registry,
		logger:   logger,
		workers:  make(map[string]*worker),
	}
	for _, device := range registry.Enabled() {
		if _, err := supervisor.build(device); err != nil {
			return nil, err
		}
	}
	return supervisor, nil
}

func (s *Supervisor) build(device devices.Device) (*Service, error) {
	cfg := deviceConfig(s.base, device)
	service, err := NewService(&cfg, s.logger)
	if err != nil {
		return nil, fmt.Errorf("device %s: %w", device.ID, err)
	}
	service.WithDeviceIdentity(device.ID, device.Name)

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.workers[device.ID]; exists {
		return nil, fmt.Errorf("%w: %s", devices.ErrDuplicateID, device.ID)
	}
	s.workers[device.ID] = &worker{device: device, service: service}
	return service, nil
}

func (s *Supervisor) Run(ctx context.Context) error {
	s.mu.RLock()
	items := make([]*worker, 0, len(s.workers))
	for _, item := range s.workers {
		items = append(items, item)
	}
	s.mu.RUnlock()

	if len(items) == 0 {
		s.logger.Printf("supervisor has no enabled devices")
	}
	sort.Slice(items, func(i, j int) bool { return items[i].device.ID < items[j].device.ID })
	for _, item := range items {
		s.logger.Printf("supervisor starting device=%s target=%s schedule=%s",
			item.device.ID, item.device.Target, item.service.Schedule().String())
		item.service.Start(ctx)
	}

	<-ctx.Done()
	return s.stopAll()
}

func (s *Supervisor) StartDevice(ctx context.Context, id string) error {
	device, err := s.registry.Get(id)
	if err != nil {
		return err
	}
	service, err := s.build(device)
	if err != nil {
		return err
	}
	s.logger.Printf("supervisor starting device=%s target=%s", device.ID, device.Target)
	service.Start(ctx)
	return nil
}

func (s *Supervisor) Primary() (*Service, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]string, 0, len(s.workers))
	for id := range s.workers {
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, false
	}
	sort.Strings(ids)
	return s.workers[ids[0]].service, true
}

func (s *Supervisor) Devices() []devices.Device {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]devices.Device, 0, len(s.workers))
	for _, item := range s.workers {
		out = append(out, item.device)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (s *Supervisor) StopDevice(id string) error {
	s.mu.Lock()
	item, ok := s.workers[id]
	if ok {
		delete(s.workers, id)
	}
	s.mu.Unlock()
	if !ok {
		return fmt.Errorf("%w: %q", devices.ErrNotFound, id)
	}
	s.logger.Printf("supervisor stopping device=%s", id)
	return item.service.Stop()
}

func (s *Supervisor) stopAll() error {
	s.mu.Lock()
	items := make([]*worker, 0, len(s.workers))
	for id, item := range s.workers {
		items = append(items, item)
		delete(s.workers, id)
	}
	s.mu.Unlock()

	var firstErr error
	for _, item := range items {
		if err := item.service.Stop(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (s *Supervisor) Service(id string) (*Service, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.workers[id]
	if !ok {
		return nil, false
	}
	return item.service, true
}

func (s *Supervisor) Statuses() []Status {
	s.mu.RLock()
	items := make([]*worker, 0, len(s.workers))
	for _, item := range s.workers {
		items = append(items, item)
	}
	s.mu.RUnlock()

	out := make([]Status, 0, len(items))
	for _, item := range items {
		out = append(out, item.service.Status())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DeviceID < out[j].DeviceID })
	return out
}

func (s *Supervisor) Fleet() FleetStatus {
	return SummarizeFleet(s.Statuses())
}

// SummarizeFleet counts a set of device statuses into one fleet view.
//
// It is separate from the supervisor because a gateway running a single
// device has no supervisor at all, and the control plane still has to answer
// the same question. Counting in one place keeps the single-device answer and
// the inventory answer from drifting apart.
func SummarizeFleet(statuses []Status) FleetStatus {
	fleet := FleetStatus{Devices: len(statuses), Statuses: statuses}

	for _, status := range statuses {
		if status.Running {
			fleet.Running++
		}
		switch health.State(status.Health.State) {
		case health.StateOnline:
			fleet.Online++
		case health.StateDegraded:
			fleet.Degraded++
		case health.StateOffline:
			fleet.Offline++
		default:
			fleet.Unknown++
		}
		switch clock.State(status.Clock.State) {
		case clock.StateOK:
			fleet.ClockOK++
		case clock.StateWarn:
			fleet.ClockWarn++
		case clock.StateCritical:
			fleet.ClockCritical++
		default:
			fleet.ClockUnknown++
		}
		if abs := absoluteDuration(status.Clock.Skew); abs > absoluteDuration(fleet.WorstClockSkew) {
			fleet.WorstClockSkew = status.Clock.Skew
			fleet.WorstClockID = status.DeviceID
		}
	}
	return fleet
}

func deviceConfig(base config.GatewayConfig, device devices.Device) config.GatewayConfig {
	cfg := base
	cfg.Target = device.Target
	cfg.CRCMode = device.ChecksumMode
	cfg.AdapterAddr = device.AdapterAddr
	cfg.ScheduleMode = device.ScheduleMode
	cfg.PollInterval = device.PollInterval
	cfg.PollOffset = device.PollOffset
	cfg.RequestTimeout = device.RequestTimeout
	cfg.ConnectTimeout = device.ConnectTimeout
	cfg.ClockWarn = device.ClockWarn
	cfg.ClockCritical = device.ClockCritical
	cfg.DegradeAfter = device.DegradeAfter
	cfg.OfflineAfter = device.OfflineAfter
	cfg.RecoverAfter = device.RecoverAfter
	cfg.GRPCListen = ""
	cfg.DevicesFile = ""
	cfg.Normalize()
	return cfg
}

func absoluteDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}
