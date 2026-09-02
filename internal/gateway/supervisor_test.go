package gateway

import (
	"context"
	"errors"
	"io"
	"log"
	"testing"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/clock"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/devices"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/health"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/schedule"
)

func testDevice(id, target string, enabled bool) devices.Device {
	return devices.Device{
		ID:             id,
		Target:         target,
		ChecksumMode:   "sum",
		AdapterAddr:    1,
		PollInterval:   40 * time.Millisecond,
		RequestTimeout: 300 * time.Millisecond,
		ConnectTimeout: 500 * time.Millisecond,
		Enabled:        enabled,
	}
}

func testRegistry(t *testing.T, items ...devices.Device) *devices.Registry {
	t.Helper()
	registry, err := devices.FromSlice(items)
	if err != nil {
		t.Fatal(err)
	}
	return registry
}

func newTestSupervisor(t *testing.T, registry *devices.Registry) *Supervisor {
	t.Helper()
	base := testGatewayConfig("127.0.0.1:1")
	supervisor, err := NewSupervisor(base, registry, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	return supervisor
}

func TestNewSupervisorRejectsMissingDependencies(t *testing.T) {
	logger := log.New(io.Discard, "", 0)

	if _, err := NewSupervisor(nil, devices.NewRegistry(), logger); err == nil {
		t.Fatal("NewSupervisor() must reject a nil base config")
	}
	if _, err := NewSupervisor(testGatewayConfig("127.0.0.1:1"), nil, logger); err == nil {
		t.Fatal("NewSupervisor() must reject a nil registry")
	}
}

func TestSupervisorBuildsOnlyEnabledDevices(t *testing.T) {
	registry := testRegistry(t,
		testDevice("beta", "127.0.0.1:9001", true),
		testDevice("alpha", "127.0.0.1:9002", true),
		testDevice("gamma", "127.0.0.1:9003", false),
	)
	supervisor := newTestSupervisor(t, registry)

	list := supervisor.Devices()
	if len(list) != 2 {
		t.Fatalf("Devices() = %d entries, want 2", len(list))
	}
	if list[0].ID != "alpha" || list[1].ID != "beta" {
		t.Fatalf("Devices() = %+v, want a sorted list", list)
	}

	if _, ok := supervisor.Service("gamma"); ok {
		t.Fatal("a disabled device must not get a worker")
	}
	if _, ok := supervisor.Service("alpha"); !ok {
		t.Fatal("an enabled device must get a worker")
	}
}

func TestSupervisorPrimaryIsTheFirstDeviceByID(t *testing.T) {
	registry := testRegistry(t,
		testDevice("zeta", "127.0.0.1:9001", true),
		testDevice("alpha", "127.0.0.1:9002", true),
	)
	supervisor := newTestSupervisor(t, registry)

	primary, ok := supervisor.Primary()
	if !ok {
		t.Fatal("Primary() must resolve when devices are enabled")
	}
	if primary.DeviceID() != "alpha" {
		t.Fatalf("Primary().DeviceID() = %q, want %q", primary.DeviceID(), "alpha")
	}
}

func TestSupervisorPrimaryWithoutDevices(t *testing.T) {
	supervisor := newTestSupervisor(t, devices.NewRegistry())

	if _, ok := supervisor.Primary(); ok {
		t.Fatal("Primary() must not resolve for an empty inventory")
	}
	if got := supervisor.Statuses(); len(got) != 0 {
		t.Fatalf("Statuses() = %d entries, want 0", len(got))
	}

	fleet := supervisor.Fleet()
	if fleet.Devices != 0 || fleet.Running != 0 {
		t.Fatalf("Fleet() = %+v, want an empty fleet", fleet)
	}
}

func TestSupervisorRunPollsEveryEnabledDevice(t *testing.T) {
	wire := testWire(t, "sum")
	handler := func(request frame.Frame, index int) deviceReply {
		return timeReply(t, wire, request, time.Now())
	}
	first := startScriptedDevice(t, "sum", handler)
	second := startScriptedDevice(t, "sum", handler)

	registry := testRegistry(t,
		testDevice("alpha", first.addr(), true),
		testDevice("beta", second.addr(), true),
	)
	supervisor := newTestSupervisor(t, registry)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- supervisor.Run(ctx) }()

	waitFor(t, 5*time.Second, func() bool {
		statuses := supervisor.Statuses()
		if len(statuses) != 2 {
			return false
		}
		for _, status := range statuses {
			if status.SuccessfulReads == 0 {
				return false
			}
		}
		return true
	})

	fleet := supervisor.Fleet()
	if fleet.Devices != 2 || fleet.Running != 2 {
		t.Fatalf("Fleet() = %+v, want two running devices", fleet)
	}
	if fleet.Online != 2 {
		t.Fatalf("Fleet().Online = %d, want 2", fleet.Online)
	}
	if fleet.Statuses[0].DeviceID != "alpha" || fleet.Statuses[1].DeviceID != "beta" {
		t.Fatalf("Fleet().Statuses = %+v, want sorted device ids", fleet.Statuses)
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Run() = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("supervisor did not stop")
	}

	if got := supervisor.Statuses(); len(got) != 0 {
		t.Fatalf("Statuses() after stop = %d entries, want 0", len(got))
	}
}

func TestSupervisorFleetTracksWorstClockSkew(t *testing.T) {
	wire := testWire(t, "sum")
	healthy := startScriptedDevice(t, "sum", func(request frame.Frame, index int) deviceReply {
		return timeReply(t, wire, request, time.Now())
	})
	drifting := startScriptedDevice(t, "sum", func(request frame.Frame, index int) deviceReply {
		return timeReply(t, wire, request, time.Now().Add(90*time.Second))
	})

	registry := testRegistry(t,
		testDevice("good", healthy.addr(), true),
		testDevice("skewed", drifting.addr(), true),
	)
	supervisor := newTestSupervisor(t, registry)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() { errCh <- supervisor.Run(ctx) }()

	waitFor(t, 5*time.Second, func() bool {
		for _, status := range supervisor.Statuses() {
			if status.DeviceID == "skewed" && status.Clock.Samples >= 3 {
				return true
			}
		}
		return false
	})

	fleet := supervisor.Fleet()
	if fleet.WorstClockID != "skewed" {
		t.Fatalf("WorstClockID = %q, want %q", fleet.WorstClockID, "skewed")
	}
	if fleet.WorstClockSkew < 80*time.Second {
		t.Fatalf("WorstClockSkew = %s, want about 90s", fleet.WorstClockSkew)
	}
	if fleet.ClockCritical == 0 {
		t.Fatalf("Fleet() = %+v, want a critical clock", fleet)
	}

	cancel()
	select {
	case <-errCh:
	case <-time.After(5 * time.Second):
		t.Fatal("supervisor did not stop")
	}
}

func TestSupervisorStartAndStopDevice(t *testing.T) {
	wire := testWire(t, "sum")
	device := startScriptedDevice(t, "sum", func(request frame.Frame, index int) deviceReply {
		return timeReply(t, wire, request, time.Now())
	})

	registry := testRegistry(t, testDevice("alpha", device.addr(), false))
	supervisor := newTestSupervisor(t, registry)

	if _, ok := supervisor.Service("alpha"); ok {
		t.Fatal("a disabled device must not be running")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := supervisor.StartDevice(ctx, "alpha"); err != nil {
		t.Fatal(err)
	}
	service, ok := supervisor.Service("alpha")
	if !ok {
		t.Fatal("StartDevice() must register a worker")
	}
	waitFor(t, 5*time.Second, func() bool { return service.Status().SuccessfulReads > 0 })

	if err := supervisor.StopDevice("alpha"); err != nil {
		t.Fatal(err)
	}
	if _, ok := supervisor.Service("alpha"); ok {
		t.Fatal("StopDevice() must drop the worker")
	}
}

func TestSupervisorStartDeviceRejectsUnknownAndDuplicate(t *testing.T) {
	registry := testRegistry(t, testDevice("alpha", "127.0.0.1:9001", true))
	supervisor := newTestSupervisor(t, registry)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := supervisor.StartDevice(ctx, "absent"); !errors.Is(err, devices.ErrNotFound) {
		t.Fatalf("StartDevice() error = %v, want %v", err, devices.ErrNotFound)
	}
	if err := supervisor.StartDevice(ctx, "alpha"); !errors.Is(err, devices.ErrDuplicateID) {
		t.Fatalf("StartDevice() error = %v, want %v", err, devices.ErrDuplicateID)
	}
	if err := supervisor.StopDevice("absent"); !errors.Is(err, devices.ErrNotFound) {
		t.Fatalf("StopDevice() error = %v, want %v", err, devices.ErrNotFound)
	}

	if err := supervisor.StopDevice("alpha"); err != nil {
		t.Fatal(err)
	}
}

func TestNewServiceRejectsInvalidSettings(t *testing.T) {
	logger := log.New(io.Discard, "", 0)

	badChecksum := testGatewayConfig("127.0.0.1:1")
	badChecksum.CRCMode = "md5"
	if _, err := NewService(badChecksum, logger); err == nil {
		t.Fatal("NewService() must reject an unknown checksum mode")
	}

	badSchedule := testGatewayConfig("127.0.0.1:1")
	badSchedule.ScheduleMode = "cron"
	if _, err := NewService(badSchedule, logger); err == nil {
		t.Fatal("NewService() must reject an unknown schedule mode")
	}

	badOffset := testGatewayConfig("127.0.0.1:1")
	badOffset.ScheduleMode = string(schedule.ModeAligned)
	badOffset.PollInterval = time.Second
	badOffset.PollOffset = 5 * time.Second
	if _, err := NewService(badOffset, logger); err == nil {
		t.Fatal("NewService() must reject an offset beyond the interval")
	}

	badRetry := testGatewayConfig("127.0.0.1:1")
	badRetry.RetryAttempts = -1
	if err := badRetry.Validate(); err == nil {
		t.Fatal("config validation must reject a negative retry count")
	}
	service, err := NewService(badRetry, logger)
	if err != nil {
		t.Fatalf("NewService() must stay defensive and clamp instead of failing: %v", err)
	}
	if got := service.Status().Retry.Attempts; got != 0 {
		t.Fatalf("Retry.Attempts = %d, want 0 after clamping", got)
	}
}

func TestDeviceConfigOverridesBaseSettings(t *testing.T) {
	base := *testGatewayConfig("127.0.0.1:1")
	base.GRPCListen = ":9200"
	base.DevicesFile = "devices.json"

	device := devices.Device{
		ID:             "alpha",
		Target:         "10.0.0.1:9000",
		ChecksumMode:   "crc16",
		AdapterAddr:    7,
		ScheduleMode:   string(schedule.ModeAligned),
		PollInterval:   time.Minute,
		PollOffset:     5 * time.Second,
		RequestTimeout: 2 * time.Second,
		ConnectTimeout: 3 * time.Second,
		ClockWarn:      time.Second,
		ClockCritical:  time.Minute,
		DegradeAfter:   2,
		OfflineAfter:   6,
		RecoverAfter:   1,
		Enabled:        true,
	}.Normalize()

	cfg := deviceConfig(base, device)

	if cfg.Target != device.Target || cfg.CRCMode != device.ChecksumMode || cfg.AdapterAddr != device.AdapterAddr {
		t.Fatalf("deviceConfig() transport = %+v", cfg)
	}
	if cfg.ScheduleMode != device.ScheduleMode || cfg.PollInterval != device.PollInterval || cfg.PollOffset != device.PollOffset {
		t.Fatalf("deviceConfig() schedule = %+v", cfg)
	}
	if cfg.ClockWarn != device.ClockWarn || cfg.ClockCritical != device.ClockCritical {
		t.Fatalf("deviceConfig() clock = (%s, %s)", cfg.ClockWarn, cfg.ClockCritical)
	}
	if cfg.DegradeAfter != device.DegradeAfter || cfg.OfflineAfter != device.OfflineAfter || cfg.RecoverAfter != device.RecoverAfter {
		t.Fatalf("deviceConfig() health = %+v", cfg.HealthPolicy())
	}
	if cfg.GRPCListen != "" {
		t.Fatal("per-device workers must not bind the shared gRPC listener")
	}
	if cfg.DevicesFile != "" {
		t.Fatal("per-device workers must not recurse into the inventory")
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("deviceConfig() must produce a valid config: %v", err)
	}
	if base.Target != "127.0.0.1:1" || base.GRPCListen != ":9200" {
		t.Fatal("deviceConfig() must not mutate the base config")
	}
}

func TestAbsoluteDuration(t *testing.T) {
	tests := []struct {
		in   time.Duration
		want time.Duration
	}{
		{in: 0, want: 0},
		{in: 5 * time.Second, want: 5 * time.Second},
		{in: -5 * time.Second, want: 5 * time.Second},
	}

	for _, tt := range tests {
		if got := absoluteDuration(tt.in); got != tt.want {
			t.Fatalf("absoluteDuration(%s) = %s, want %s", tt.in, got, tt.want)
		}
	}
}

func TestFleetCountsUnknownStates(t *testing.T) {
	registry := testRegistry(t, testDevice("alpha", "127.0.0.1:9001", true))
	supervisor := newTestSupervisor(t, registry)

	fleet := supervisor.Fleet()
	if fleet.Devices != 1 {
		t.Fatalf("Fleet().Devices = %d, want 1", fleet.Devices)
	}
	if fleet.Unknown != 1 {
		t.Fatalf("Fleet().Unknown = %d, want 1", fleet.Unknown)
	}
	if fleet.ClockUnknown != 1 {
		t.Fatalf("Fleet().ClockUnknown = %d, want 1", fleet.ClockUnknown)
	}
	if fleet.Running != 0 {
		t.Fatalf("Fleet().Running = %d, want 0", fleet.Running)
	}

	status := fleet.Statuses[0]
	if status.Health.State != string(health.StateUnknown) {
		t.Fatalf("Health.State = %q", status.Health.State)
	}
	if status.Clock.State != string(clock.StateUnknown) {
		t.Fatalf("Clock.State = %q", status.Clock.State)
	}
}

func TestSupervisorRunWithoutDevices(t *testing.T) {
	supervisor := newTestSupervisor(t, devices.NewRegistry())

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- supervisor.Run(ctx) }()

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Run() = %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("supervisor did not stop")
	}
}
