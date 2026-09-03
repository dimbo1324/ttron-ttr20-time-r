package gateway

import (
	"errors"
	"fmt"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/clock"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/health"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/retry"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/schedule"
)

// ErrInvalidSettings is the sentinel every rejected update wraps, so a caller
// can tell "you asked for something impossible" apart from "something broke".
var ErrInvalidSettings = errors.New("invalid gateway settings")

// Settings is the part of a gateway's configuration that can change while it
// is running: how often it polls, how hard it tries, and where the lines are
// that turn a measurement into an alarm.
//
// What is deliberately absent is as much of the design as what is here.
//
// The target address, checksum mode and adapter address are not settings but
// identity: change one mid-flight and the frames on the wire stop matching the
// device the gateway is connected to. That is a different gateway, and asking
// for it means restarting one, not tuning this one.
//
// The clock and health window sizes are absent too. Resizing a ring buffer
// discards the history in it, so a operator nudging a window would silently
// erase the very evidence they were looking at.
type Settings struct {
	ScheduleMode   string
	PollInterval   time.Duration
	PollOffset     time.Duration
	RequestTimeout time.Duration
	RetryAttempts  int
	RetryDelay     time.Duration
	ClockWarn      time.Duration
	ClockCritical  time.Duration
	DegradeAfter   int
	OfflineAfter   int
	RecoverAfter   int
}

// Schedule builds the plan these settings describe, and is the same call the
// service makes when applying them -- so a Settings that validates is a
// Settings that can be installed.
func (s Settings) Schedule() (schedule.Schedule, error) {
	mode, err := schedule.ParseMode(s.ScheduleMode)
	if err != nil {
		return nil, err
	}
	return schedule.New(mode, s.PollInterval, s.PollOffset)
}

func (s Settings) RetryPolicy() retry.Policy {
	return retry.Policy{Attempts: s.RetryAttempts, Delay: s.RetryDelay}
}

func (s Settings) Thresholds() clock.Thresholds {
	return clock.Thresholds{Warn: s.ClockWarn, Critical: s.ClockCritical}
}

// Policy carries no window size: the tracker keeps the one it was built with,
// because changing it would throw away the outcomes already in the window.
func (s Settings) Policy() health.Policy {
	return health.Policy{
		DegradeAfter: s.DegradeAfter,
		OfflineAfter: s.OfflineAfter,
		RecoverAfter: s.RecoverAfter,
	}
}

// Validate checks the request as it was written, without repairing it.
//
// The same choice the gateway config makes: normalising a wrong value into a
// working one hides the mistake, and an operator who typed a 30ms interval
// deserves to be told rather than quietly given 5s.
func (s Settings) Validate() error {
	if _, err := s.Schedule(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidSettings, err)
	}
	if s.RequestTimeout <= 0 {
		return fmt.Errorf("%w: request timeout must be positive", ErrInvalidSettings)
	}
	// A request that may outlast the interval it is scheduled in cannot be a
	// schedule; it is a queue that grows until the line gives up.
	if s.RequestTimeout >= s.PollInterval {
		return fmt.Errorf("%w: request timeout must be below the poll interval", ErrInvalidSettings)
	}
	if err := s.RetryPolicy().Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidSettings, err)
	}
	if err := s.Thresholds().Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidSettings, err)
	}
	if err := s.Policy().Normalize().Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidSettings, err)
	}
	if s.DegradeAfter <= 0 || s.OfflineAfter <= 0 || s.RecoverAfter <= 0 {
		return fmt.Errorf("%w: health thresholds must be positive", ErrInvalidSettings)
	}
	if s.OfflineAfter < s.DegradeAfter {
		return fmt.Errorf("%w: offline threshold must not be below the degrade threshold", ErrInvalidSettings)
	}
	return nil
}

// Settings reports what the gateway is running with now.
func (s *Service) Settings() Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings
}

// UpdateSettings installs a new configuration on a gateway that may be
// mid-poll, and returns what it is now running with.
//
// All or nothing: everything is validated before anything is applied, because
// a half-applied update leaves a gateway in a state no one asked for and no
// one can name. Nothing here interrupts an exchange already in flight -- the
// request timeout a poll started under is the one it finishes under.
func (s *Service) UpdateSettings(next Settings) (Settings, error) {
	if err := next.Validate(); err != nil {
		return Settings{}, err
	}
	plan, err := next.Schedule()
	if err != nil {
		return Settings{}, fmt.Errorf("%w: %w", ErrInvalidSettings, err)
	}
	policy := next.RetryPolicy().Normalize()
	if err := policy.Validate(); err != nil {
		return Settings{}, fmt.Errorf("%w: %w", ErrInvalidSettings, err)
	}

	// The clock monitor and the health tracker own their own locks and their
	// own normalisation, so they are told directly rather than through the
	// service's state.
	thresholds := s.skew.SetThresholds(next.Thresholds())
	health := s.health.SetPolicy(next.Policy())

	s.mu.Lock()
	rescheduled := s.settings.ScheduleMode != next.ScheduleMode ||
		s.settings.PollInterval != next.PollInterval ||
		s.settings.PollOffset != next.PollOffset

	next.ClockWarn = thresholds.Warn
	next.ClockCritical = thresholds.Critical
	next.DegradeAfter = health.DegradeAfter
	next.OfflineAfter = health.OfflineAfter
	next.RecoverAfter = health.RecoverAfter
	next.RetryAttempts = policy.Attempts
	next.RetryDelay = policy.Delay

	s.settings = next
	s.schedule = plan
	s.retry = policy
	s.status.PollingInterval = next.PollInterval
	s.status.RequestTimeout = next.RequestTimeout
	s.status.Schedule.Mode = string(plan.Mode())
	s.status.Schedule.Interval = plan.Interval()
	s.status.Schedule.Offset = next.PollOffset
	s.status.Schedule.Description = plan.String()
	s.status.Retry.Attempts = policy.Attempts
	s.status.Retry.Delay = policy.Delay
	s.status.Retry.MaxDelay = policy.MaxDelay
	s.mu.Unlock()

	s.logger.Printf("gateway settings updated schedule=%s timeout=%s retries=%d warn=%s critical=%s health=%d/%d/%d",
		plan.String(), next.RequestTimeout, policy.Attempts,
		thresholds.Warn, thresholds.Critical,
		health.DegradeAfter, health.OfflineAfter, health.RecoverAfter)

	// A session parked on a one-minute wait must not sit there for a minute
	// after being told to poll every second; see waitForPoll.
	if rescheduled {
		s.signalReschedule()
	}
	return next, nil
}

// signalReschedule wakes a waiting session without blocking the caller. The
// channel holds one token: a second update arriving before the session woke up
// needs no second wake-up.
func (s *Service) signalReschedule() {
	select {
	case s.reschedule <- struct{}{}:
	default:
	}
}

/* --------------------------------------------- what the poll loop reads */

// The poll loop runs on its own goroutine and these values can change under
// it, so every one of them is read through a lock rather than off the struct.

func (s *Service) currentSchedule() schedule.Schedule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.schedule
}

func (s *Service) currentRetry() retry.Policy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.retry
}

func (s *Service) requestTimeout() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings.RequestTimeout
}
