package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/clock"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/health"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/retry"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/schedule"
)

type GatewayConfig struct {
	Target         string
	CRCMode        string
	AdapterAddr    int
	PollInterval   time.Duration
	RequestTimeout time.Duration
	ConnectTimeout time.Duration
	BackoffInitial time.Duration
	BackoffMax     time.Duration
	RecentSize     int
	LogFile        string
	GRPCListen     string

	ScheduleMode  string
	PollOffset    time.Duration
	RetryAttempts int
	RetryDelay    time.Duration

	ClockWarn       time.Duration
	ClockCritical   time.Duration
	ClockWindowSize int

	DegradeAfter     int
	OfflineAfter     int
	RecoverAfter     int
	HealthWindowSize int

	IdentityProbe bool
	DevicesFile   string
}

// DefaultGateway is the assignment this project was written for: read the
// device clock on the fifth second of every minute.
//
// A general polling service could reasonably default to "every five seconds"
// and leave the calendar to the operator. This one does not, because the
// behaviour someone gets from `go run ./cmd/ft12-gateway` with no flags is the
// behaviour they will believe the program has -- and a repository that answers
// a specific brief should answer it when run, not only when configured.
//
// The cost is a sharp edge worth knowing about: an aligned schedule needs its
// offset to fall inside its interval, so shortening the interval alone can
// make the pair invalid. Validate says so by name; see below.
func DefaultGateway() GatewayConfig {
	return GatewayConfig{
		Target:         "127.0.0.1:9000",
		CRCMode:        "sum",
		AdapterAddr:    1,
		PollInterval:   time.Minute,
		RequestTimeout: 1500 * time.Millisecond,
		ConnectTimeout: 2 * time.Second,
		BackoffInitial: 500 * time.Millisecond,
		BackoffMax:     5 * time.Second,
		RecentSize:     100,
		LogFile:        "runtime/logs/ft12-gateway.log",
		GRPCListen:     ":9200",

		ScheduleMode:  string(schedule.ModeAligned),
		PollOffset:    5 * time.Second,
		RetryAttempts: retry.DefaultAttempts,
		RetryDelay:    retry.DefaultDelay,

		ClockWarn:       clock.DefaultWarnThreshold,
		ClockCritical:   clock.DefaultCriticalThreshold,
		ClockWindowSize: clock.DefaultWindowSize,

		DegradeAfter:     health.DefaultDegradeAfter,
		OfflineAfter:     health.DefaultOfflineAfter,
		RecoverAfter:     health.DefaultRecoverAfter,
		HealthWindowSize: health.DefaultWindowSize,

		IdentityProbe: true,
		DevicesFile:   "",
	}
}

func LoadGateway(args []string) (*GatewayConfig, error) {
	c := DefaultGateway()
	fs := flag.NewFlagSet("ft12-gateway", flag.ContinueOnError)
	fs.StringVar(&c.Target, "target", c.Target, "target emulator/device TCP address")
	fs.StringVar(&c.CRCMode, "crc", c.CRCMode, "crc mode: sum | crc16")
	mode := fs.String("mode", "", "checksum mode alias: sum | crc16")
	fs.IntVar(&c.AdapterAddr, "adapter", c.AdapterAddr, "adapter address byte (0..255)")
	fs.DurationVar(&c.PollInterval, "interval", c.PollInterval, "polling interval")
	fs.DurationVar(&c.RequestTimeout, "timeout", c.RequestTimeout, "request/response timeout")
	fs.DurationVar(&c.ConnectTimeout, "connect-timeout", c.ConnectTimeout, "TCP connect timeout")
	fs.DurationVar(&c.BackoffInitial, "backoff-initial", c.BackoffInitial, "initial reconnect backoff")
	fs.DurationVar(&c.BackoffMax, "backoff-max", c.BackoffMax, "maximum reconnect backoff")
	fs.IntVar(&c.RecentSize, "recent", c.RecentSize, "recent frame/event buffer size")
	fs.StringVar(&c.LogFile, "log", c.LogFile, "path to log file; empty = stdout")
	fs.StringVar(&c.GRPCListen, "grpc-listen", c.GRPCListen, "gRPC control listen address; empty disables gRPC")
	fs.StringVar(&c.ScheduleMode, "schedule", c.ScheduleMode, "poll schedule mode: interval | aligned")
	fs.DurationVar(&c.PollOffset, "poll-offset", c.PollOffset, "offset inside the aligned poll interval")
	fs.IntVar(&c.RetryAttempts, "retry-attempts", c.RetryAttempts, "in-session retries after a protocol error or timeout")
	fs.DurationVar(&c.RetryDelay, "retry-delay", c.RetryDelay, "initial delay between in-session retries")
	fs.DurationVar(&c.ClockWarn, "clock-warn", c.ClockWarn, "absolute device clock skew that raises a warning")
	fs.DurationVar(&c.ClockCritical, "clock-critical", c.ClockCritical, "absolute device clock skew that raises a critical alarm")
	fs.IntVar(&c.ClockWindowSize, "clock-window", c.ClockWindowSize, "number of skew samples kept for median and drift")
	fs.IntVar(&c.DegradeAfter, "degrade-after", c.DegradeAfter, "consecutive failures before the device is degraded")
	fs.IntVar(&c.OfflineAfter, "offline-after", c.OfflineAfter, "consecutive failures before the device is offline")
	fs.IntVar(&c.RecoverAfter, "recover-after", c.RecoverAfter, "consecutive successes before the device recovers")
	fs.IntVar(&c.HealthWindowSize, "health-window", c.HealthWindowSize, "number of poll outcomes kept for availability and latency")
	fs.BoolVar(&c.IdentityProbe, "identity-probe", c.IdentityProbe, "read device identity once per connection")
	fs.StringVar(&c.DevicesFile, "devices", c.DevicesFile, "device inventory JSON file; empty polls the single -target device")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if *mode != "" {
		c.CRCMode = *mode
	}
	c.Normalize()
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

func MustLoadGatewayFromOS() *GatewayConfig {
	cfg, err := LoadGateway(os.Args[1:])
	if err != nil {
		panic(err)
	}
	return cfg
}

func (c *GatewayConfig) Normalize() {
	if c.CRCMode == "" {
		c.CRCMode = "sum"
	}
	if c.ScheduleMode == "" {
		c.ScheduleMode = string(schedule.ModeInterval)
	}
	if c.ClockWindowSize == 0 {
		c.ClockWindowSize = clock.DefaultWindowSize
	}
	if c.HealthWindowSize == 0 {
		c.HealthWindowSize = health.DefaultWindowSize
	}
	if c.ClockWarn == 0 {
		c.ClockWarn = clock.DefaultWarnThreshold
	}
	if c.ClockCritical == 0 {
		c.ClockCritical = clock.DefaultCriticalThreshold
	}
	if c.DegradeAfter == 0 {
		c.DegradeAfter = health.DefaultDegradeAfter
	}
	if c.OfflineAfter == 0 {
		c.OfflineAfter = health.DefaultOfflineAfter
	}
	if c.RecoverAfter == 0 {
		c.RecoverAfter = health.DefaultRecoverAfter
	}
}

func (c GatewayConfig) ClockThresholds() clock.Thresholds {
	return clock.Thresholds{Warn: c.ClockWarn, Critical: c.ClockCritical}
}

func (c GatewayConfig) HealthPolicy() health.Policy {
	return health.Policy{
		DegradeAfter: c.DegradeAfter,
		OfflineAfter: c.OfflineAfter,
		RecoverAfter: c.RecoverAfter,
		WindowSize:   c.HealthWindowSize,
	}
}

func (c GatewayConfig) RetryPolicy() retry.Policy {
	return retry.Policy{Attempts: c.RetryAttempts, Delay: c.RetryDelay}
}

func (c GatewayConfig) Schedule() (schedule.Schedule, error) {
	mode, err := schedule.ParseMode(c.ScheduleMode)
	if err != nil {
		return nil, err
	}
	return schedule.New(mode, c.PollInterval, c.PollOffset)
}

func (c GatewayConfig) Validate() error {
	if c.Target == "" {
		return fmt.Errorf("target must not be empty")
	}
	if err := validateTCPAddress(c.Target, "target address"); err != nil {
		return err
	}
	if err := validateChecksumMode(c.CRCMode); err != nil {
		return err
	}
	if err := validateAdapterAddr(c.AdapterAddr); err != nil {
		return err
	}
	if err := validatePositiveDuration(c.PollInterval, "poll interval"); err != nil {
		return err
	}
	if err := validatePositiveDuration(c.RequestTimeout, "request timeout"); err != nil {
		return err
	}
	if err := validatePositiveDuration(c.ConnectTimeout, "connect timeout"); err != nil {
		return err
	}
	if err := validatePositiveDuration(c.BackoffInitial, "backoff initial"); err != nil {
		return err
	}
	if err := validatePositiveDuration(c.BackoffMax, "backoff max"); err != nil {
		return err
	}
	if c.BackoffInitial > c.BackoffMax {
		return fmt.Errorf("backoff initial must not exceed backoff max")
	}
	if err := validateRecentSize(c.RecentSize); err != nil {
		return err
	}
	if c.GRPCListen != "" {
		if err := validateTCPAddress(c.GRPCListen, "gRPC listen address"); err != nil {
			return err
		}
	}
	if _, err := c.Schedule(); err != nil {
		// Named by flag rather than by field. The default schedule is aligned
		// at +5s, so the first thing anyone does -- pass -interval on its own
		// -- is also the first way to make the pair invalid, and the bare
		// "offset must be below the interval" leaves them hunting for a knob
		// they never touched.
		if errors.Is(err, schedule.ErrInvalidOffset) {
			return fmt.Errorf("%w: -poll-offset %s does not fit inside -interval %s (use -schedule interval for a fixed rate)",
				err, c.PollOffset, c.PollInterval)
		}
		return err
	}
	if err := c.RetryPolicy().Validate(); err != nil {
		return err
	}
	if err := c.ClockThresholds().Validate(); err != nil {
		return err
	}
	if err := c.HealthPolicy().Validate(); err != nil {
		return err
	}
	if err := validateRecentSize(c.ClockWindowSize); err != nil {
		return fmt.Errorf("clock window: %w", err)
	}
	if err := validateRecentSize(c.HealthWindowSize); err != nil {
		return fmt.Errorf("health window: %w", err)
	}
	return nil
}
