package devices

import (
	"errors"
	"testing"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/clock"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/health"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/schedule"
)

func validDevice() Device {
	return Device{
		ID:      "tekon-01",
		Target:  "127.0.0.1:9000",
		Enabled: true,
	}
}

func TestDeviceNormalizeFillsDefaults(t *testing.T) {
	got := Device{ID: "  TEKON-01 ", Target: " 127.0.0.1:9000 "}.Normalize()

	if got.ID != "tekon-01" {
		t.Fatalf("ID = %q, want %q", got.ID, "tekon-01")
	}
	if got.Target != "127.0.0.1:9000" {
		t.Fatalf("Target = %q", got.Target)
	}
	if got.Name != "tekon-01" {
		t.Fatalf("Name = %q, want the id as fallback", got.Name)
	}
	if got.ChecksumMode != "sum" {
		t.Fatalf("ChecksumMode = %q", got.ChecksumMode)
	}
	if got.ScheduleMode != string(schedule.ModeInterval) {
		t.Fatalf("ScheduleMode = %q", got.ScheduleMode)
	}
	if got.PollInterval != DefaultPollInterval {
		t.Fatalf("PollInterval = %s", got.PollInterval)
	}
	if got.RequestTimeout != DefaultRequestTimeout || got.ConnectTimeout != DefaultConnectTimeout {
		t.Fatalf("timeouts = (%s, %s)", got.RequestTimeout, got.ConnectTimeout)
	}
	if got.ClockWarn != clock.DefaultWarnThreshold || got.ClockCritical != clock.DefaultCriticalThreshold {
		t.Fatalf("clock thresholds = (%s, %s)", got.ClockWarn, got.ClockCritical)
	}
	if got.DegradeAfter != health.DefaultDegradeAfter || got.OfflineAfter != health.DefaultOfflineAfter {
		t.Fatalf("health policy = %+v", got.Policy())
	}
	if err := got.Validate(); err != nil {
		t.Fatalf("normalized device must be valid: %v", err)
	}
}

func TestDeviceNormalizeKeepsExplicitValues(t *testing.T) {
	in := Device{
		ID:             "meter-7",
		Name:           "Feeder 7",
		Target:         "10.0.0.5:9100",
		ChecksumMode:   "crc16",
		AdapterAddr:    9,
		ScheduleMode:   "aligned",
		PollInterval:   time.Minute,
		PollOffset:     5 * time.Second,
		RequestTimeout: 2 * time.Second,
		ConnectTimeout: 3 * time.Second,
		ClockWarn:      time.Second,
		ClockCritical:  time.Minute,
		DegradeAfter:   2,
		OfflineAfter:   6,
		RecoverAfter:   1,
	}

	got := in.Normalize()
	if got.Name != "Feeder 7" || got.PollInterval != time.Minute || got.PollOffset != 5*time.Second {
		t.Fatalf("Normalize() = %+v", got)
	}
	if got.DegradeAfter != 2 || got.OfflineAfter != 6 || got.RecoverAfter != 1 {
		t.Fatalf("health policy = %+v", got.Policy())
	}
	if err := got.Validate(); err != nil {
		t.Fatalf("Validate() = %v", err)
	}
}

func TestDeviceNormalizeClampsNegativeOffset(t *testing.T) {
	got := Device{ID: "a", Target: "127.0.0.1:1", PollOffset: -time.Second}.Normalize()
	if got.PollOffset != 0 {
		t.Fatalf("PollOffset = %s, want 0", got.PollOffset)
	}
}

func TestDeviceValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Device)
		wantErr error
	}{
		{name: "valid", mutate: func(*Device) {}},
		{name: "empty id", mutate: func(d *Device) { d.ID = "" }, wantErr: ErrInvalidID},
		{name: "id with spaces", mutate: func(d *Device) { d.ID = "te kon" }, wantErr: ErrInvalidID},
		{name: "id with slash", mutate: func(d *Device) { d.ID = "a/b" }, wantErr: ErrInvalidID},
		{name: "empty target", mutate: func(d *Device) { d.Target = "" }, wantErr: ErrInvalidTarget},
		{name: "target without port", mutate: func(d *Device) { d.Target = "127.0.0.1" }, wantErr: ErrInvalidTarget},
		{name: "target with bad port", mutate: func(d *Device) { d.Target = "127.0.0.1:abc" }, wantErr: ErrInvalidTarget},
		{name: "target with port zero", mutate: func(d *Device) { d.Target = "127.0.0.1:0" }, wantErr: ErrInvalidTarget},
		{name: "target with port above range", mutate: func(d *Device) { d.Target = "127.0.0.1:70000" }, wantErr: ErrInvalidTarget},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := validDevice()
			tt.mutate(&device)
			device = device.Normalize()
			err := device.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeviceIDCanonicalization(t *testing.T) {
	raw := Device{ID: "TEKON-01", Target: "127.0.0.1:9000"}
	if err := raw.Validate(); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("Validate() on a raw upper-case id = %v, want %v", err, ErrInvalidID)
	}
	if err := raw.Normalize().Validate(); err != nil {
		t.Fatalf("Normalize() must canonicalize the id: %v", err)
	}
}

func TestDeviceValidateRejectsBadFields(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Device)
	}{
		{name: "unknown checksum mode", mutate: func(d *Device) { d.ChecksumMode = "md5" }},
		{name: "adapter below range", mutate: func(d *Device) { d.AdapterAddr = -1 }},
		{name: "adapter above range", mutate: func(d *Device) { d.AdapterAddr = 256 }},
		{name: "unknown schedule mode", mutate: func(d *Device) { d.ScheduleMode = "cron" }},
		{name: "offset beyond interval", mutate: func(d *Device) {
			d.ScheduleMode = "aligned"
			d.PollInterval = time.Second
			d.PollOffset = 5 * time.Second
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := validDevice()
			tt.mutate(&device)
			device = device.Normalize()
			if err := device.Validate(); err == nil {
				t.Fatalf("Validate() = nil, want an error for %+v", device)
			}
		})
	}
}

func TestDeviceValidateRejectsNonPositiveDurations(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Device)
	}{
		{name: "poll interval", mutate: func(d *Device) { d.PollInterval = -time.Second }},
		{name: "request timeout", mutate: func(d *Device) { d.RequestTimeout = -time.Second }},
		{name: "connect timeout", mutate: func(d *Device) { d.ConnectTimeout = -time.Second }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := validDevice()
			tt.mutate(&device)
			if err := device.Validate(); err == nil {
				t.Fatalf("Validate() = nil, want an error")
			}
		})
	}
}

func TestDeviceSchedule(t *testing.T) {
	device := Device{
		ID:           "a",
		Target:       "127.0.0.1:1",
		ScheduleMode: "aligned",
		PollInterval: time.Minute,
		PollOffset:   5 * time.Second,
	}.Normalize()

	plan, err := device.Schedule()
	if err != nil {
		t.Fatal(err)
	}
	if plan.Mode() != schedule.ModeAligned || plan.Interval() != time.Minute {
		t.Fatalf("Schedule() = %s", plan.String())
	}

	broken := device
	broken.ScheduleMode = "cron"
	if _, err := broken.Schedule(); err == nil {
		t.Fatal("Schedule() must reject an unknown mode")
	}
}

func TestSplitHostPort(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		wantHost string
		wantPort int
		wantErr  bool
	}{
		{name: "ipv4", address: "127.0.0.1:9000", wantHost: "127.0.0.1", wantPort: 9000},
		{name: "ipv6", address: "[::1]:9000", wantHost: "::1", wantPort: 9000},
		{name: "hostname", address: "emulator:9000", wantHost: "emulator", wantPort: 9000},
		{name: "missing port", address: "127.0.0.1", wantErr: true},
		{name: "non numeric port", address: "127.0.0.1:port", wantErr: true},
		{name: "port zero", address: "127.0.0.1:0", wantErr: true},
		{name: "port too large", address: "127.0.0.1:65536", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port, err := splitHostPort(tt.address)
			if (err != nil) != tt.wantErr {
				t.Fatalf("splitHostPort(%q) error = %v, wantErr %t", tt.address, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if host != tt.wantHost || port != tt.wantPort {
				t.Fatalf("splitHostPort(%q) = (%q, %d), want (%q, %d)", tt.address, host, port, tt.wantHost, tt.wantPort)
			}
		})
	}
}
