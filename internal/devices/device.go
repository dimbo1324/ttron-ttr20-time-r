package devices

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/clock"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/health"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/schedule"
)

var (
	ErrInvalidID     = errors.New("invalid device id")
	ErrInvalidTarget = errors.New("invalid device target")
	ErrDuplicateID   = errors.New("duplicate device id")
	ErrNotFound      = errors.New("device not found")
)

var idPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)

const (
	DefaultPollInterval   = 5 * time.Second
	DefaultRequestTimeout = 1500 * time.Millisecond
	DefaultConnectTimeout = 2 * time.Second
)

type Device struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Target         string        `json:"target"`
	ChecksumMode   string        `json:"checksumMode"`
	AdapterAddr    int           `json:"adapterAddr"`
	ScheduleMode   string        `json:"scheduleMode"`
	PollInterval   time.Duration `json:"pollInterval"`
	PollOffset     time.Duration `json:"pollOffset"`
	RequestTimeout time.Duration `json:"requestTimeout"`
	ConnectTimeout time.Duration `json:"connectTimeout"`
	ClockWarn      time.Duration `json:"clockWarn"`
	ClockCritical  time.Duration `json:"clockCritical"`
	DegradeAfter   int           `json:"degradeAfter"`
	OfflineAfter   int           `json:"offlineAfter"`
	RecoverAfter   int           `json:"recoverAfter"`
	Enabled        bool          `json:"enabled"`
}

func (d Device) Normalize() Device {
	d.ID = strings.ToLower(strings.TrimSpace(d.ID))
	d.Target = strings.TrimSpace(d.Target)
	d.Name = strings.TrimSpace(d.Name)
	if d.Name == "" {
		d.Name = d.ID
	}
	if d.ChecksumMode == "" {
		d.ChecksumMode = string(checksum.ModeSum)
	}
	if d.ScheduleMode == "" {
		d.ScheduleMode = string(schedule.ModeInterval)
	}
	if d.PollInterval <= 0 {
		d.PollInterval = DefaultPollInterval
	}
	if d.RequestTimeout <= 0 {
		d.RequestTimeout = DefaultRequestTimeout
	}
	if d.ConnectTimeout <= 0 {
		d.ConnectTimeout = DefaultConnectTimeout
	}
	if d.PollOffset < 0 {
		d.PollOffset = 0
	}
	thresholds := d.Thresholds()
	d.ClockWarn = thresholds.Warn
	d.ClockCritical = thresholds.Critical
	policy := d.Policy()
	d.DegradeAfter = policy.DegradeAfter
	d.OfflineAfter = policy.OfflineAfter
	d.RecoverAfter = policy.RecoverAfter
	return d
}

func (d Device) Thresholds() clock.Thresholds {
	return clock.Thresholds{Warn: d.ClockWarn, Critical: d.ClockCritical}.Normalize()
}

func (d Device) Policy() health.Policy {
	return health.Policy{
		DegradeAfter: d.DegradeAfter,
		OfflineAfter: d.OfflineAfter,
		RecoverAfter: d.RecoverAfter,
	}.Normalize()
}

func (d Device) Schedule() (schedule.Schedule, error) {
	mode, err := schedule.ParseMode(d.ScheduleMode)
	if err != nil {
		return nil, err
	}
	return schedule.New(mode, d.PollInterval, d.PollOffset)
}

func (d Device) Validate() error {
	if !idPattern.MatchString(d.ID) {
		return fmt.Errorf("%w: %q must match %s", ErrInvalidID, d.ID, idPattern.String())
	}
	if d.Target == "" {
		return fmt.Errorf("%w: target must not be empty", ErrInvalidTarget)
	}
	if _, _, err := splitHostPort(d.Target); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidTarget, err)
	}
	if _, err := checksum.ParseMode(d.ChecksumMode); err != nil {
		return err
	}
	if d.AdapterAddr < 0 || d.AdapterAddr > 255 {
		return fmt.Errorf("device %q: adapter address must be in range 0..255", d.ID)
	}
	if d.PollInterval <= 0 {
		return fmt.Errorf("device %q: poll interval must be positive", d.ID)
	}
	if d.RequestTimeout <= 0 {
		return fmt.Errorf("device %q: request timeout must be positive", d.ID)
	}
	if d.ConnectTimeout <= 0 {
		return fmt.Errorf("device %q: connect timeout must be positive", d.ID)
	}
	if _, err := d.Schedule(); err != nil {
		return fmt.Errorf("device %q: %w", d.ID, err)
	}
	if err := d.Thresholds().Validate(); err != nil {
		return fmt.Errorf("device %q: %w", d.ID, err)
	}
	if err := d.Policy().Validate(); err != nil {
		return fmt.Errorf("device %q: %w", d.ID, err)
	}
	return nil
}
