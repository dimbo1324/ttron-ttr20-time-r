package gateway

import (
	"context"
	"testing"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/clock"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/health"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/observability/events"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/command"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/schedule"
)

func TestServiceExposesItsConfiguration(t *testing.T) {
	cfg := testGatewayConfig("127.0.0.1:9000")
	cfg.ScheduleMode = string(schedule.ModeAligned)
	cfg.PollInterval = time.Minute
	cfg.PollOffset = 5 * time.Second
	service := newTestService(t, cfg)

	if got := service.Schedule().Mode(); got != schedule.ModeAligned {
		t.Fatalf("Schedule().Mode() = %q", got)
	}
	if !service.Commands().Supports(command.ReadTime) || !service.Commands().Supports(command.ReadIdentity) {
		t.Fatal("the service must expose the default command registry")
	}

	status := service.Status()
	if status.TargetAddress != cfg.Target || status.ChecksumMode != "sum" {
		t.Fatalf("Status() = %+v", status)
	}
	if status.Schedule.Description == "" {
		t.Fatal("Status().Schedule.Description must be populated")
	}
	if status.Retry.Attempts != cfg.RetryAttempts || status.Retry.Delay != cfg.RetryDelay {
		t.Fatalf("Status().Retry = %+v", status.Retry)
	}
	if !status.Identity.Supported {
		t.Fatal("identity support must start optimistic")
	}
	if status.Clock.State != string(clock.StateUnknown) || status.Health.State != string(health.StateUnknown) {
		t.Fatalf("initial states = (%q, %q)", status.Clock.State, status.Health.State)
	}
}

func TestServiceWithDeviceIdentity(t *testing.T) {
	service := newTestService(t, testGatewayConfig("127.0.0.1:9000"))

	returned := service.WithDeviceIdentity("meter-1", "Feeder 1")
	if returned != service {
		t.Fatal("WithDeviceIdentity() must return the same service for chaining")
	}
	if service.DeviceID() != "meter-1" {
		t.Fatalf("DeviceID() = %q", service.DeviceID())
	}

	status := service.Status()
	if status.DeviceID != "meter-1" || status.DeviceName != "Feeder 1" {
		t.Fatalf("Status() = %+v", status)
	}
}

func TestServiceSetClockThresholds(t *testing.T) {
	service := newTestService(t, testGatewayConfig("127.0.0.1:9000"))

	applied := service.SetClockThresholds(clock.Thresholds{Warn: time.Second, Critical: time.Minute})
	if applied.Warn != time.Second || applied.Critical != time.Minute {
		t.Fatalf("SetClockThresholds() = %+v", applied)
	}
	status := service.Status()
	if status.Clock.WarnThreshold != time.Second || status.Clock.CriticalThreshold != time.Minute {
		t.Fatalf("Status().Clock = %+v", status.Clock)
	}

	normalized := service.SetClockThresholds(clock.Thresholds{})
	if normalized != clock.DefaultThresholds() {
		t.Fatalf("SetClockThresholds() = %+v, want defaults", normalized)
	}
}

func TestServiceClockReportAndHealthSnapshot(t *testing.T) {
	service := newTestService(t, testGatewayConfig("127.0.0.1:9000"))

	if service.ClockReport().Observed() {
		t.Fatal("ClockReport() must start empty")
	}
	if got := service.HealthSnapshot().State; got != health.StateUnknown {
		t.Fatalf("HealthSnapshot().State = %q", got)
	}

	moment := time.Now()
	service.observeClock(clock.Sample{
		RequestedAt: moment,
		ReceivedAt:  moment.Add(10 * time.Millisecond),
		DeviceTime:  moment.Add(time.Minute),
	})
	service.observeHealthSuccess(10 * time.Millisecond)

	report := service.ClockReport()
	if !report.Observed() || report.State != clock.StateCritical {
		t.Fatalf("ClockReport() = %+v", report)
	}
	if got := service.HealthSnapshot().State; got != health.StateOnline {
		t.Fatalf("HealthSnapshot().State = %q", got)
	}
}

func TestServiceRejectsInvalidClockSample(t *testing.T) {
	service := newTestService(t, testGatewayConfig("127.0.0.1:9000"))

	service.observeClock(clock.Sample{})

	status := service.Status()
	if status.Clock.RejectedSamples != 1 {
		t.Fatalf("Clock.RejectedSamples = %d, want 1", status.Clock.RejectedSamples)
	}
	if status.Clock.State != string(clock.StateUnknown) {
		t.Fatalf("Clock.State = %q, want unknown", status.Clock.State)
	}
}

func TestServiceRecordsSystemEvents(t *testing.T) {
	service := newTestService(t, testGatewayConfig("127.0.0.1:9000"))

	service.recordSystemEvent("custom", "manual entry", time.Time{})
	service.recordIdentityEvent(command.Identity{Model: "TTR20", Serial: "SN-1", Firmware: "1.0"}, time.Now())

	snapshot := service.Snapshot()
	if len(snapshot.Recent) != 2 {
		t.Fatalf("Snapshot().Recent = %d entries, want 2", len(snapshot.Recent))
	}
	for _, record := range snapshot.Recent {
		if record.Direction != events.DirectionSystem {
			t.Fatalf("record direction = %q, want %q", record.Direction, events.DirectionSystem)
		}
		if record.Timestamp.IsZero() {
			t.Fatal("a system event must carry a timestamp")
		}
		if record.Service != events.ServiceGateway {
			t.Fatalf("record service = %q", record.Service)
		}
	}
	if snapshot.Status.RecentFramesCount != 2 {
		t.Fatalf("RecentFramesCount = %d, want 2", snapshot.Status.RecentFramesCount)
	}
}

func TestServiceRecordsClockAndHealthTransitions(t *testing.T) {
	wire := testWire(t, "sum")
	device := startScriptedDevice(t, "sum", func(request frame.Frame, index int) deviceReply {
		return timeReply(t, wire, request, time.Now().Add(90*time.Second))
	})

	service := newTestService(t, testGatewayConfig(device.addr()))
	runService(t, service)

	waitFor(t, 3*time.Second, func() bool { return service.Status().Clock.Samples >= 1 })

	var sawClockEvent, sawDeviceEvent bool
	for _, record := range service.Snapshot().Recent {
		switch record.Command {
		case "clock-state":
			sawClockEvent = true
		case "device-state":
			sawDeviceEvent = true
		}
	}
	if !sawClockEvent {
		t.Fatal("a clock state change must be recorded in the event ring")
	}
	if !sawDeviceEvent {
		t.Fatal("a device state change must be recorded in the event ring")
	}
}

func TestServiceStartIsIdempotentAndStopIsSafe(t *testing.T) {
	wire := testWire(t, "sum")
	device := startScriptedDevice(t, "sum", func(request frame.Frame, index int) deviceReply {
		return timeReply(t, wire, request, time.Now())
	})

	service := newTestService(t, testGatewayConfig(device.addr()))

	if err := service.Stop(); err != nil {
		t.Fatalf("Stop() before Start() = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	service.Start(ctx)
	service.Start(ctx)

	waitFor(t, 3*time.Second, func() bool { return service.Status().SuccessfulReads > 0 })

	if err := service.Stop(); err != nil {
		t.Fatalf("Stop() = %v", err)
	}
	if status := service.Status(); status.Running || status.Connected {
		t.Fatalf("Status() after stop = %+v", status)
	}
	if err := service.Stop(); err != nil {
		t.Fatalf("second Stop() = %v", err)
	}
}

func TestServiceStopsWhileWaitingForNextPoll(t *testing.T) {
	wire := testWire(t, "sum")
	device := startScriptedDevice(t, "sum", func(request frame.Frame, index int) deviceReply {
		return timeReply(t, wire, request, time.Now())
	})

	cfg := testGatewayConfig(device.addr())
	cfg.PollInterval = time.Hour
	service := newTestService(t, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	service.Start(ctx)
	waitFor(t, 3*time.Second, func() bool { return service.Status().SuccessfulReads > 0 })

	cancel()
	done := make(chan error, 1)
	go func() { done <- service.Stop() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Stop() = %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Stop() must interrupt the poll wait")
	}
}

func TestServiceReconnectsAfterUnreachableTarget(t *testing.T) {
	cfg := testGatewayConfig("127.0.0.1:1")
	cfg.ConnectTimeout = 100 * time.Millisecond
	service := newTestService(t, cfg)
	runService(t, service)

	waitFor(t, 5*time.Second, func() bool { return service.Status().ConnectionAttempts >= 2 })

	status := service.Status()
	if status.Connected {
		t.Fatal("an unreachable target must not report a connection")
	}
	if status.FailedReads == 0 || status.LastError == "" {
		t.Fatalf("Status() = %+v", status)
	}
	if status.Health.State == string(health.StateOnline) {
		t.Fatalf("Health.State = %q, want a non-online state", status.Health.State)
	}
}

func TestSleepContext(t *testing.T) {
	if !sleepContext(context.Background(), 0) {
		t.Fatal("sleepContext() with a zero delay must continue")
	}

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if sleepContext(cancelled, 0) {
		t.Fatal("sleepContext() must report a cancelled context even with a zero delay")
	}
	if sleepContext(cancelled, time.Hour) {
		t.Fatal("sleepContext() must return early on cancellation")
	}

	start := time.Now()
	if !sleepContext(context.Background(), 20*time.Millisecond) {
		t.Fatal("sleepContext() must complete a short sleep")
	}
	if elapsed := time.Since(start); elapsed < 10*time.Millisecond {
		t.Fatalf("sleepContext() returned after %s", elapsed)
	}
}
