package gateway

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/clock"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/health"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/observability/events"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/codec"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/command"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/retry"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/schedule"
)

type Service struct {
	cfg    *config.GatewayConfig
	mode   checksum.Mode
	wire   codec.Codec
	logger *log.Logger

	history  *events.Ring
	schedule schedule.Schedule
	retry    retry.Policy
	commands *command.Registry
	skew     *clock.Monitor
	health   *health.Tracker

	deviceID   string
	deviceName string

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
	plan, err := cfg.Schedule()
	if err != nil {
		return nil, err
	}
	retryPolicy := cfg.RetryPolicy().Normalize()
	if err := retryPolicy.Validate(); err != nil {
		return nil, err
	}

	s := &Service{
		cfg:      cfg,
		mode:     mode,
		wire:     codec.New(mode, 0x00, byte(cfg.AdapterAddr&0xFF)),
		logger:   logger,
		history:  events.NewRing(cfg.RecentSize),
		schedule: plan,
		retry:    retryPolicy,
		commands: command.DefaultRegistry(),
		skew:     clock.NewMonitor(cfg.ClockThresholds(), cfg.ClockWindowSize),
		health:   health.NewTracker(cfg.HealthPolicy()),
	}
	s.status = Status{
		TargetAddress:   cfg.Target,
		ChecksumMode:    string(mode),
		PollingInterval: cfg.PollInterval,
		RequestTimeout:  cfg.RequestTimeout,
		ConnectTimeout:  cfg.ConnectTimeout,
		Schedule: ScheduleStatus{
			Mode:        string(plan.Mode()),
			Interval:    plan.Interval(),
			Offset:      cfg.PollOffset,
			Description: plan.String(),
		},
		Retry: RetryStatus{
			Attempts: retryPolicy.Attempts,
			Delay:    retryPolicy.Delay,
			MaxDelay: retryPolicy.MaxDelay,
		},
		Identity: IdentityStatus{Supported: true},
	}
	return s, nil
}

func (s *Service) WithDeviceIdentity(id, name string) *Service {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deviceID = id
	s.deviceName = name
	s.status.DeviceID = id
	s.status.DeviceName = name
	return s
}

func (s *Service) Start(ctx context.Context) {
	s.runMu.Lock()
	defer s.runMu.Unlock()
	if s.cancel != nil {
		s.logger.Printf("gateway polling already running")
		return
	}
	s.logger.Printf("gateway polling start schedule=%s retries=%d", s.schedule.String(), s.retry.Attempts)
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
	out := s.status
	s.mu.RUnlock()

	out.RecentFramesCount = s.history.Len()
	out.Clock = s.clockStatus()
	out.Health = s.healthStatus()
	return out
}

func (s *Service) Snapshot() Snapshot {
	return Snapshot{Status: s.Status(), Recent: s.history.Snapshot()}
}

func (s *Service) Schedule() schedule.Schedule {
	return s.schedule
}

func (s *Service) Commands() *command.Registry {
	return s.commands
}

func (s *Service) ClockReport() clock.Report {
	return s.skew.Report()
}

func (s *Service) SetClockThresholds(thresholds clock.Thresholds) clock.Thresholds {
	applied := s.skew.SetThresholds(thresholds)
	s.logger.Printf("gateway clock thresholds updated warn=%s critical=%s", applied.Warn, applied.Critical)
	return applied
}

func (s *Service) HealthSnapshot() health.Snapshot {
	return s.health.Snapshot()
}

// History is the raw material behind the two aggregates -- the skew window and
// the outcome window, oldest first. Status reports what those windows add up
// to; this reports what is in them, for a caller that has to draw the run.
func (s *Service) History() History {
	return History{
		ClockSamples:   s.skew.Samples(),
		HealthOutcomes: s.health.Outcomes(),
	}
}

func (s *Service) DeviceID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.deviceID
}

func (s *Service) clockStatus() ClockStatus {
	report := s.skew.Report()
	observed, rejected := s.skew.Counters()
	thresholds := s.skew.Thresholds()

	status := ClockStatus{
		State:             string(s.skew.State()),
		WarnThreshold:     thresholds.Warn,
		CriticalThreshold: thresholds.Critical,
		ObservedSamples:   observed,
		RejectedSamples:   rejected,
	}
	if !report.Observed() {
		return status
	}
	status.Skew = report.Skew
	status.MedianSkew = report.MedianSkew
	status.MinSkew = report.MinSkew
	status.MaxSkew = report.MaxSkew
	status.DriftPerDay = report.Drift.PerDay
	status.DriftDetermined = report.Drift.Determined
	status.DriftFit = report.Drift.Fit
	status.Samples = report.SampleCount
	status.RoundTrip = report.RoundTrip
	status.UpdatedAt = report.At
	return status
}

func (s *Service) healthStatus() HealthStatus {
	snapshot := s.health.Snapshot()
	return HealthStatus{
		State:                string(snapshot.State),
		Since:                snapshot.Since,
		Availability:         snapshot.Availability,
		WindowSamples:        snapshot.WindowSamples,
		ConsecutiveFailures:  snapshot.ConsecutiveFailures,
		ConsecutiveSuccesses: snapshot.ConsecutiveSuccesses,
		LatencyP50:           snapshot.Latency.P50,
		LatencyP95:           snapshot.Latency.P95,
		LatencyP99:           snapshot.Latency.P99,
		LatencyMax:           snapshot.Latency.Max,
		LatencyMean:          snapshot.Latency.Mean,
		DegradeAfter:         snapshot.Policy.DegradeAfter,
		OfflineAfter:         snapshot.Policy.OfflineAfter,
		RecoverAfter:         snapshot.Policy.RecoverAfter,
	}
}

func (s *Service) nextPollAt(at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.Schedule.NextPollAt = at
}
