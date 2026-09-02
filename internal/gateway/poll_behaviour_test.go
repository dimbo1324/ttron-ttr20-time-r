package gateway

import (
	"testing"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/clock"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/health"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/command"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/schedule"
)

func TestGatewayRetriesProtocolErrorWithoutReconnecting(t *testing.T) {
	const corrupted = 2
	wire := testWire(t, "sum")

	device := startScriptedDevice(t, "sum", func(request frame.Frame, index int) deviceReply {
		if index < corrupted {
			return corruptTimeReply(t, wire, request, time.Now())
		}
		return timeReply(t, wire, request, time.Now())
	})

	service := newTestService(t, testGatewayConfig(device.addr()))
	runService(t, service)

	waitFor(t, 3*time.Second, func() bool { return service.Status().SuccessfulReads > 0 })

	status := service.Status()
	if status.ReconnectCount != 0 {
		t.Fatalf("ReconnectCount = %d, want 0: a bad frame must not drop the connection", status.ReconnectCount)
	}
	if status.Retry.TotalRetries < corrupted {
		t.Fatalf("Retry.TotalRetries = %d, want at least %d", status.Retry.TotalRetries, corrupted)
	}
	if status.ProtocolErrors == 0 {
		t.Fatal("ProtocolErrors must count the corrupted frames")
	}
	if device.connectionCount() != 1 {
		t.Fatalf("device saw %d connections, want 1", device.connectionCount())
	}
}

func TestGatewayExhaustsRetriesAndKeepsConnection(t *testing.T) {
	wire := testWire(t, "sum")
	device := startScriptedDevice(t, "sum", func(request frame.Frame, index int) deviceReply {
		return corruptTimeReply(t, wire, request, time.Now())
	})

	service := newTestService(t, testGatewayConfig(device.addr()))
	runService(t, service)

	waitFor(t, 3*time.Second, func() bool { return service.Status().Retry.ExhaustedPolls > 0 })

	status := service.Status()
	if status.SuccessfulReads != 0 {
		t.Fatalf("SuccessfulReads = %d, want 0", status.SuccessfulReads)
	}
	if status.FailedReads == 0 {
		t.Fatal("FailedReads must be recorded once retries are exhausted")
	}
	if status.ReconnectCount != 0 {
		t.Fatalf("ReconnectCount = %d, want 0", status.ReconnectCount)
	}
	if !status.Connected {
		t.Fatal("the connection must stay open after a protocol failure")
	}
	if device.connectionCount() != 1 {
		t.Fatalf("device saw %d connections, want 1", device.connectionCount())
	}
}

func TestGatewayReconnectsWhenDeviceClosesConnection(t *testing.T) {
	wire := testWire(t, "sum")
	device := startScriptedDevice(t, "sum", func(request frame.Frame, index int) deviceReply {
		reply := timeReply(t, wire, request, time.Now())
		reply.close = true
		return reply
	})

	service := newTestService(t, testGatewayConfig(device.addr()))
	runService(t, service)

	waitFor(t, 3*time.Second, func() bool { return service.Status().ReconnectCount > 0 })
	waitFor(t, 3*time.Second, func() bool { return device.connectionCount() >= 2 })
}

func TestGatewayRetriesTimeoutOnSameConnection(t *testing.T) {
	wire := testWire(t, "sum")
	device := startScriptedDevice(t, "sum", func(request frame.Frame, index int) deviceReply {
		if index == 0 {
			return deviceReply{}
		}
		return timeReply(t, wire, request, time.Now())
	})

	cfg := testGatewayConfig(device.addr())
	cfg.RequestTimeout = 120 * time.Millisecond
	service := newTestService(t, cfg)
	runService(t, service)

	waitFor(t, 3*time.Second, func() bool { return service.Status().SuccessfulReads > 0 })

	status := service.Status()
	if status.Retry.TotalRetries == 0 {
		t.Fatal("a silent response must be retried")
	}
	if status.ReconnectCount != 0 {
		t.Fatalf("ReconnectCount = %d, want 0", status.ReconnectCount)
	}
}

func TestGatewayRecordsClockSkewAndState(t *testing.T) {
	const offset = 45 * time.Second
	wire := testWire(t, "sum")

	device := startScriptedDevice(t, "sum", func(request frame.Frame, index int) deviceReply {
		time.Sleep(2 * time.Millisecond)
		return timeReply(t, wire, request, time.Now().Add(offset))
	})

	cfg := testGatewayConfig(device.addr())
	cfg.ClockWarn = 2 * time.Second
	cfg.ClockCritical = 30 * time.Second
	service := newTestService(t, cfg)
	runService(t, service)

	waitFor(t, 3*time.Second, func() bool { return service.Status().Clock.Samples >= 3 })

	status := service.Status()
	if status.Clock.State != string(clock.StateCritical) {
		t.Fatalf("Clock.State = %q, want %q", status.Clock.State, clock.StateCritical)
	}
	if delta := status.Clock.Skew - offset; delta > time.Second || delta < -time.Second {
		t.Fatalf("Clock.Skew = %s, want about %s", status.Clock.Skew, offset)
	}
	if status.Clock.WarnThreshold != cfg.ClockWarn || status.Clock.CriticalThreshold != cfg.ClockCritical {
		t.Fatalf("thresholds = (%s, %s)", status.Clock.WarnThreshold, status.Clock.CriticalThreshold)
	}
	if status.Clock.ObservedSamples == 0 || status.Clock.UpdatedAt.IsZero() {
		t.Fatalf("Clock = %+v", status.Clock)
	}
	if status.LastRoundTrip <= 0 {
		t.Fatalf("LastRoundTrip = %s, want a positive round trip", status.LastRoundTrip)
	}
	if status.Clock.RoundTrip <= 0 {
		t.Fatalf("Clock.RoundTrip = %s, want a positive round trip", status.Clock.RoundTrip)
	}
}

func TestGatewayKeepsClockHealthyForSyncedDevice(t *testing.T) {
	wire := testWire(t, "sum")
	device := startScriptedDevice(t, "sum", func(request frame.Frame, index int) deviceReply {
		return timeReply(t, wire, request, time.Now())
	})

	service := newTestService(t, testGatewayConfig(device.addr()))
	runService(t, service)

	waitFor(t, 3*time.Second, func() bool { return service.Status().Clock.Samples >= 3 })

	status := service.Status()
	if status.Clock.State != string(clock.StateOK) {
		t.Fatalf("Clock.State = %q, want %q", status.Clock.State, clock.StateOK)
	}
	if status.Health.State != string(health.StateOnline) {
		t.Fatalf("Health.State = %q, want %q", status.Health.State, health.StateOnline)
	}
	if status.Health.Availability != 1 {
		t.Fatalf("Health.Availability = %f, want 1", status.Health.Availability)
	}
	if status.Health.WindowSamples < 3 {
		t.Fatalf("Health.WindowSamples = %d, want at least 3", status.Health.WindowSamples)
	}
	if status.Health.LatencyP50 < 0 || status.Health.LatencyMax < status.Health.LatencyP50 {
		t.Fatalf("latency percentiles are inconsistent: %+v", status.Health)
	}
}

func TestGatewayMarksDeviceOfflineAfterRepeatedFailures(t *testing.T) {
	device := startScriptedDevice(t, "sum", func(request frame.Frame, index int) deviceReply {
		return deviceReply{}
	})

	cfg := testGatewayConfig(device.addr())
	cfg.RequestTimeout = 60 * time.Millisecond
	cfg.RetryAttempts = 0
	cfg.DegradeAfter = 2
	cfg.OfflineAfter = 3
	cfg.RecoverAfter = 1
	service := newTestService(t, cfg)
	runService(t, service)

	waitFor(t, 5*time.Second, func() bool {
		return service.Status().Health.State == string(health.StateOffline)
	})

	status := service.Status()
	if status.Health.ConsecutiveFailures < 3 {
		t.Fatalf("ConsecutiveFailures = %d, want at least 3", status.Health.ConsecutiveFailures)
	}
	if status.Health.Availability != 0 {
		t.Fatalf("Availability = %f, want 0", status.Health.Availability)
	}
	if status.LastError == "" {
		t.Fatal("LastError must describe the failure")
	}
}

func TestGatewayAlignedScheduleDoesNotPollOnConnect(t *testing.T) {
	wire := testWire(t, "sum")
	device := startScriptedDevice(t, "sum", func(request frame.Frame, index int) deviceReply {
		return timeReply(t, wire, request, time.Now())
	})

	cfg := testGatewayConfig(device.addr())
	cfg.ScheduleMode = string(schedule.ModeAligned)
	cfg.PollInterval = time.Hour
	cfg.PollOffset = 5 * time.Second
	service := newTestService(t, cfg)
	runService(t, service)

	waitFor(t, 2*time.Second, func() bool { return service.Status().Connected })
	time.Sleep(200 * time.Millisecond)

	status := service.Status()
	if status.SuccessfulReads != 0 || device.requestCount() != 0 {
		t.Fatalf("aligned mode must wait for the next boundary: reads=%d requests=%d",
			status.SuccessfulReads, device.requestCount())
	}
	if status.Schedule.Mode != string(schedule.ModeAligned) {
		t.Fatalf("Schedule.Mode = %q", status.Schedule.Mode)
	}
	if status.Schedule.Offset != 5*time.Second || status.Schedule.Interval != time.Hour {
		t.Fatalf("Schedule = %+v", status.Schedule)
	}
	if status.Schedule.NextPollAt.IsZero() {
		t.Fatal("Schedule.NextPollAt must be published")
	}
	if status.Schedule.NextPollAt.Second() != 5 {
		t.Fatalf("NextPollAt = %s, want the fifth second", status.Schedule.NextPollAt)
	}
}

func TestGatewayIntervalSchedulePollsOnConnect(t *testing.T) {
	wire := testWire(t, "sum")
	device := startScriptedDevice(t, "sum", func(request frame.Frame, index int) deviceReply {
		return timeReply(t, wire, request, time.Now())
	})

	cfg := testGatewayConfig(device.addr())
	cfg.PollInterval = time.Hour
	service := newTestService(t, cfg)
	runService(t, service)

	waitFor(t, 3*time.Second, func() bool { return service.Status().SuccessfulReads == 1 })
}

func TestGatewayIdentityProbeRecordsDevice(t *testing.T) {
	wire := testWire(t, "sum")
	want := command.Identity{Model: "TTR20", Serial: "SN-0000042", Firmware: "1.2.3"}

	device := startScriptedDevice(t, "sum", func(request frame.Frame, index int) deviceReply {
		if requestedCommand(request) == command.ReadIdentity {
			return identityReply(t, wire, request, want)
		}
		return timeReply(t, wire, request, time.Now())
	})

	cfg := testGatewayConfig(device.addr())
	cfg.IdentityProbe = true
	service := newTestService(t, cfg)
	runService(t, service)

	waitFor(t, 3*time.Second, func() bool { return service.Status().Identity.Known })

	identity := service.Status().Identity
	if identity.Model != want.Model || identity.Serial != want.Serial || identity.Firmware != want.Firmware {
		t.Fatalf("Identity = %+v, want %+v", identity, want)
	}
	if !identity.Supported || identity.ReadAt.IsZero() {
		t.Fatalf("Identity = %+v", identity)
	}
}

func TestGatewayIdentityProbeToleratesUnsupportedDevice(t *testing.T) {
	wire := testWire(t, "sum")
	device := startScriptedDevice(t, "sum", func(request frame.Frame, index int) deviceReply {
		if requestedCommand(request) == command.ReadIdentity {
			return ackReply(t, wire, request)
		}
		return timeReply(t, wire, request, time.Now())
	})

	cfg := testGatewayConfig(device.addr())
	cfg.IdentityProbe = true
	service := newTestService(t, cfg)
	runService(t, service)

	waitFor(t, 3*time.Second, func() bool { return service.Status().SuccessfulReads > 0 })

	identity := service.Status().Identity
	if identity.Known {
		t.Fatalf("Identity = %+v, want unknown", identity)
	}
	if identity.Supported {
		t.Fatal("an unsupported device must be remembered as unsupported")
	}
	if service.Status().ReconnectCount != 0 {
		t.Fatal("a failed identity probe must not drop the connection")
	}
}
