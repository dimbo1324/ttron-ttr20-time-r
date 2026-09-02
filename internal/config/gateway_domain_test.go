package config

import (
	"testing"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/clock"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/health"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/retry"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/schedule"
)

func TestDefaultGatewayIsValid(t *testing.T) {
	cfg := DefaultGateway()
	cfg.Normalize()

	if err := cfg.Validate(); err != nil {
		t.Fatalf("DefaultGateway() must be valid: %v", err)
	}
	if cfg.ScheduleMode != string(schedule.ModeInterval) {
		t.Fatalf("ScheduleMode = %q", cfg.ScheduleMode)
	}
	if cfg.RetryAttempts != retry.DefaultAttempts || cfg.RetryDelay != retry.DefaultDelay {
		t.Fatalf("retry defaults = (%d, %s)", cfg.RetryAttempts, cfg.RetryDelay)
	}
	if cfg.ClockWarn != clock.DefaultWarnThreshold || cfg.ClockCritical != clock.DefaultCriticalThreshold {
		t.Fatalf("clock defaults = (%s, %s)", cfg.ClockWarn, cfg.ClockCritical)
	}
	if cfg.DegradeAfter != health.DefaultDegradeAfter || cfg.OfflineAfter != health.DefaultOfflineAfter {
		t.Fatalf("health defaults = %+v", cfg.HealthPolicy())
	}
	if !cfg.IdentityProbe {
		t.Fatal("the identity probe must be enabled by default")
	}
	if cfg.DevicesFile != "" {
		t.Fatalf("DevicesFile = %q, want empty by default", cfg.DevicesFile)
	}
}

func TestLoadGatewayParsesDomainFlags(t *testing.T) {
	cfg, err := LoadGateway([]string{
		"-target", "10.0.0.5:9000",
		"-schedule", "aligned",
		"-interval", "1m",
		"-poll-offset", "5s",
		"-retry-attempts", "4",
		"-retry-delay", "50ms",
		"-clock-warn", "3s",
		"-clock-critical", "45s",
		"-clock-window", "32",
		"-degrade-after", "2",
		"-offline-after", "8",
		"-recover-after", "3",
		"-health-window", "64",
		"-identity-probe=false",
		"-devices", "inventory.json",
	})
	if err != nil {
		t.Fatal(err)
	}

	if cfg.ScheduleMode != "aligned" || cfg.PollInterval != time.Minute || cfg.PollOffset != 5*time.Second {
		t.Fatalf("schedule = (%q, %s, %s)", cfg.ScheduleMode, cfg.PollInterval, cfg.PollOffset)
	}
	if cfg.RetryAttempts != 4 || cfg.RetryDelay != 50*time.Millisecond {
		t.Fatalf("retry = (%d, %s)", cfg.RetryAttempts, cfg.RetryDelay)
	}
	if cfg.ClockWarn != 3*time.Second || cfg.ClockCritical != 45*time.Second || cfg.ClockWindowSize != 32 {
		t.Fatalf("clock = (%s, %s, %d)", cfg.ClockWarn, cfg.ClockCritical, cfg.ClockWindowSize)
	}
	if cfg.DegradeAfter != 2 || cfg.OfflineAfter != 8 || cfg.RecoverAfter != 3 || cfg.HealthWindowSize != 64 {
		t.Fatalf("health = %+v window=%d", cfg.HealthPolicy(), cfg.HealthWindowSize)
	}
	if cfg.IdentityProbe {
		t.Fatal("-identity-probe=false must disable the probe")
	}
	if cfg.DevicesFile != "inventory.json" {
		t.Fatalf("DevicesFile = %q", cfg.DevicesFile)
	}
}

func TestGatewayNormalizeFillsDomainDefaults(t *testing.T) {
	cfg := GatewayConfig{
		Target:         "127.0.0.1:9000",
		PollInterval:   time.Second,
		RequestTimeout: time.Second,
		ConnectTimeout: time.Second,
		BackoffInitial: time.Second,
		BackoffMax:     time.Second,
		RecentSize:     10,
	}
	cfg.Normalize()

	if cfg.CRCMode != "sum" || cfg.ScheduleMode != string(schedule.ModeInterval) {
		t.Fatalf("Normalize() modes = (%q, %q)", cfg.CRCMode, cfg.ScheduleMode)
	}
	if cfg.ClockWindowSize != clock.DefaultWindowSize || cfg.HealthWindowSize != health.DefaultWindowSize {
		t.Fatalf("window sizes = (%d, %d)", cfg.ClockWindowSize, cfg.HealthWindowSize)
	}
	if cfg.ClockWarn != clock.DefaultWarnThreshold || cfg.ClockCritical != clock.DefaultCriticalThreshold {
		t.Fatalf("clock thresholds = (%s, %s)", cfg.ClockWarn, cfg.ClockCritical)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("a normalized config must be valid: %v", err)
	}
}

func TestGatewayNormalizeDoesNotRepairContradictoryValues(t *testing.T) {
	cfg := DefaultGateway()
	cfg.ClockWarn = time.Minute
	cfg.ClockCritical = time.Second
	cfg.Normalize()

	if cfg.ClockCritical != time.Second || cfg.ClockWarn != time.Minute {
		t.Fatalf("Normalize() rewrote explicit values: (%s, %s)", cfg.ClockWarn, cfg.ClockCritical)
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() must reject a critical threshold below the warn threshold")
	}
}

func TestGatewayNormalizeOnlyFillsUnsetValues(t *testing.T) {
	cfg := DefaultGateway()
	cfg.ClockWindowSize = -1
	cfg.HealthWindowSize = -1
	cfg.DegradeAfter = -1
	cfg.Normalize()

	if cfg.ClockWindowSize != -1 || cfg.HealthWindowSize != -1 || cfg.DegradeAfter != -1 {
		t.Fatal("Normalize() must leave explicitly wrong values for Validate() to report")
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() must reject negative sizes")
	}
}

func TestGatewayValidateRejectsDomainMistakes(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*GatewayConfig)
	}{
		{name: "unknown schedule mode", mutate: func(c *GatewayConfig) { c.ScheduleMode = "cron" }},
		{
			name: "offset beyond interval",
			mutate: func(c *GatewayConfig) {
				c.ScheduleMode = string(schedule.ModeAligned)
				c.PollInterval = time.Second
				c.PollOffset = 5 * time.Second
			},
		},
		{
			name: "negative offset",
			mutate: func(c *GatewayConfig) {
				c.ScheduleMode = string(schedule.ModeAligned)
				c.PollOffset = -time.Second
			},
		},
		{name: "negative retry attempts", mutate: func(c *GatewayConfig) { c.RetryAttempts = -1 }},
		{name: "too many retry attempts", mutate: func(c *GatewayConfig) { c.RetryAttempts = retry.MaxAttempts + 1 }},
		{name: "negative retry delay", mutate: func(c *GatewayConfig) { c.RetryDelay = -time.Second }},
		{name: "zero clock window", mutate: func(c *GatewayConfig) { c.ClockWindowSize = 0 }},
		{name: "negative clock window", mutate: func(c *GatewayConfig) { c.ClockWindowSize = -1 }},
		{name: "zero health window", mutate: func(c *GatewayConfig) { c.HealthWindowSize = 0 }},
		{name: "offline below degrade", mutate: func(c *GatewayConfig) { c.DegradeAfter = 9; c.OfflineAfter = 2 }},
		{name: "zero recover", mutate: func(c *GatewayConfig) { c.RecoverAfter = 0 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultGateway()
			tt.mutate(&cfg)
			if err := cfg.Validate(); err == nil {
				t.Fatalf("Validate() = nil, want an error for %+v", cfg)
			}
		})
	}
}

func TestGatewayScheduleAccessor(t *testing.T) {
	cfg := DefaultGateway()
	cfg.ScheduleMode = string(schedule.ModeAligned)
	cfg.PollInterval = time.Minute
	cfg.PollOffset = 5 * time.Second

	plan, err := cfg.Schedule()
	if err != nil {
		t.Fatal(err)
	}
	if plan.Mode() != schedule.ModeAligned || plan.Interval() != time.Minute {
		t.Fatalf("Schedule() = %s", plan.String())
	}

	next := plan.Next(time.Date(2026, 9, 2, 17, 3, 4, 0, time.UTC))
	if next.Second() != 5 {
		t.Fatalf("Next() = %s, want the fifth second", next)
	}

	cfg.ScheduleMode = "cron"
	if _, err := cfg.Schedule(); err == nil {
		t.Fatal("Schedule() must reject an unknown mode")
	}
}

func TestGatewayPolicyAccessors(t *testing.T) {
	cfg := DefaultGateway()
	cfg.ClockWarn = time.Second
	cfg.ClockCritical = time.Minute
	cfg.RetryAttempts = 3
	cfg.RetryDelay = 20 * time.Millisecond
	cfg.DegradeAfter = 2
	cfg.OfflineAfter = 5
	cfg.RecoverAfter = 1
	cfg.HealthWindowSize = 16

	if got := cfg.ClockThresholds(); got.Warn != time.Second || got.Critical != time.Minute {
		t.Fatalf("ClockThresholds() = %+v", got)
	}
	if got := cfg.RetryPolicy(); got.Attempts != 3 || got.Delay != 20*time.Millisecond {
		t.Fatalf("RetryPolicy() = %+v", got)
	}
	policy := cfg.HealthPolicy()
	if policy.DegradeAfter != 2 || policy.OfflineAfter != 5 || policy.RecoverAfter != 1 || policy.WindowSize != 16 {
		t.Fatalf("HealthPolicy() = %+v", policy)
	}
	if err := policy.Validate(); err != nil {
		t.Fatalf("HealthPolicy() must be valid: %v", err)
	}
}

func TestLoadGatewayRejectsInvalidDomainFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "unknown schedule", args: []string{"-schedule", "cron"}},
		{name: "offset beyond interval", args: []string{"-schedule", "aligned", "-interval", "1s", "-poll-offset", "5s"}},
		{name: "negative retries", args: []string{"-retry-attempts", "-1"}},
		{name: "negative clock window", args: []string{"-clock-window", "-1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := LoadGateway(tt.args); err == nil {
				t.Fatalf("LoadGateway(%v) = nil error, want a failure", tt.args)
			}
		})
	}
}

func TestLoadEmulatorParsesClockAndIdentityFlags(t *testing.T) {
	cfg, err := LoadEmulator([]string{
		"-clock-offset", "90s",
		"-clock-drift", "-12s",
		"-identity-model", "TTR20-X",
		"-identity-serial", "SN-7",
		"-identity-firmware", "9.9.9",
	})
	if err != nil {
		t.Fatal(err)
	}

	if cfg.ClockOffset != 90*time.Second || cfg.ClockDrift != -12*time.Second {
		t.Fatalf("clock faults = (%s, %s)", cfg.ClockOffset, cfg.ClockDrift)
	}
	if cfg.IdentityModel != "TTR20-X" || cfg.IdentitySerial != "SN-7" || cfg.IdentityFirmware != "9.9.9" {
		t.Fatalf("identity = (%q, %q, %q)", cfg.IdentityModel, cfg.IdentitySerial, cfg.IdentityFirmware)
	}
}

func TestEmulatorNormalizeFillsIdentityDefaults(t *testing.T) {
	cfg := DefaultEmulator()
	cfg.IdentityModel = "   "
	cfg.IdentitySerial = ""
	cfg.IdentityFirmware = "  "
	cfg.Normalize()

	if cfg.IdentityModel == "" || cfg.IdentitySerial == "" || cfg.IdentityFirmware == "" {
		t.Fatalf("identity defaults = (%q, %q, %q)", cfg.IdentityModel, cfg.IdentitySerial, cfg.IdentityFirmware)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() = %v", err)
	}
}
