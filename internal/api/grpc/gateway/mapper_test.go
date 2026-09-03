package gateway

import (
	"context"
	"io"
	"log"
	"testing"
	"time"

	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	domain "github.com/dimbo1324/ttron-ttr20-time-r/internal/gateway"
)

// richStatus is a status with every nested section populated and no two
// numbers equal, so a mapper that crosses two fields cannot pass by accident.
func richStatus() domain.Status {
	at := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	return domain.Status{
		Running:                true,
		TargetAddress:          "127.0.0.1:9000",
		ChecksumMode:           "sum",
		PollingInterval:        5 * time.Second,
		RequestTimeout:         1500 * time.Millisecond,
		ConnectTimeout:         2 * time.Second,
		Connected:              true,
		ConnectionAttempts:     11,
		SuccessfulReads:        12,
		FailedReads:            13,
		ReconnectCount:         14,
		ProtocolErrors:         15,
		LastSuccessfulReadTime: at,
		LastParsedDeviceTime:   at.Add(-time.Second),
		LastTXTimestamp:        at.Add(-2 * time.Second),
		LastRXTimestamp:        at.Add(-3 * time.Second),
		LastRoundTrip:          17 * time.Millisecond,
		RecentFramesCount:      18,
		DeviceID:               "alpha",
		DeviceName:             "Alpha meter",
		Schedule: domain.ScheduleStatus{
			Mode:        "aligned",
			Interval:    60 * time.Second,
			Offset:      5 * time.Second,
			Description: "every minute at +5s",
			NextPollAt:  at.Add(time.Minute),
		},
		Retry: domain.RetryStatus{
			Attempts:       3,
			Delay:          200 * time.Millisecond,
			MaxDelay:       2 * time.Second,
			TotalRetries:   21,
			ExhaustedPolls: 22,
		},
		Clock: domain.ClockStatus{
			State:             "warn",
			Skew:              -2500 * time.Millisecond,
			MedianSkew:        -2400 * time.Millisecond,
			MinSkew:           -3000 * time.Millisecond,
			MaxSkew:           -1800 * time.Millisecond,
			DriftPerDay:       24 * time.Second,
			DriftDetermined:   true,
			DriftFit:          0.94,
			Samples:           31,
			WarnThreshold:     2 * time.Second,
			CriticalThreshold: 30 * time.Second,
			RoundTrip:         9 * time.Millisecond,
			UpdatedAt:         at,
			ObservedSamples:   32,
			RejectedSamples:   33,
		},
		Health: domain.HealthStatus{
			State:                "degraded",
			Since:                at.Add(-time.Hour),
			Availability:         0.87,
			WindowSamples:        41,
			ConsecutiveFailures:  42,
			ConsecutiveSuccesses: 43,
			LatencyP50:           44 * time.Millisecond,
			LatencyP95:           45 * time.Millisecond,
			LatencyP99:           46 * time.Millisecond,
			LatencyMax:           47 * time.Millisecond,
			LatencyMean:          48 * time.Millisecond,
			DegradeAfter:         3,
			OfflineAfter:         10,
			RecoverAfter:         2,
		},
		Identity: domain.IdentityStatus{
			Known:     true,
			Supported: true,
			Model:     "TTR20",
			Serial:    "SN-42",
			Firmware:  "1.2.3",
			ReadAt:    at.Add(-time.Minute),
		},
	}
}

func TestMapStatusCarriesTheFlatFields(t *testing.T) {
	got := mapStatus(richStatus())

	if got.GetState() != ft12v1.ServiceState_SERVICE_STATE_RUNNING {
		t.Fatalf("State = %v", got.GetState())
	}
	if got.GetTargetAddr() != "127.0.0.1:9000" {
		t.Fatalf("TargetAddr = %q", got.GetTargetAddr())
	}
	if got.GetProtocolErrors() != 15 {
		t.Fatalf("ProtocolErrors = %d", got.GetProtocolErrors())
	}
	if got.GetLastRoundTripMs() != 17 {
		t.Fatalf("LastRoundTripMs = %d", got.GetLastRoundTripMs())
	}
	if got.GetDeviceId() != "alpha" || got.GetDeviceName() != "Alpha meter" {
		t.Fatalf("device = %q/%q", got.GetDeviceId(), got.GetDeviceName())
	}
	if got.GetPollingIntervalMs() != 5000 || got.GetRequestTimeoutMs() != 1500 {
		t.Fatalf("timings = %d/%d", got.GetPollingIntervalMs(), got.GetRequestTimeoutMs())
	}
}

func TestMapStatusCarriesTheSchedule(t *testing.T) {
	got := mapStatus(richStatus()).GetSchedule()

	if got.GetMode() != "aligned" {
		t.Fatalf("Mode = %q", got.GetMode())
	}
	if got.GetIntervalMs() != 60_000 || got.GetOffsetMs() != 5000 {
		t.Fatalf("interval/offset = %d/%d", got.GetIntervalMs(), got.GetOffsetMs())
	}
	if got.GetNextPollAt() == nil {
		t.Fatal("NextPollAt must be carried")
	}
}

func TestMapStatusCarriesTheRetryBudget(t *testing.T) {
	got := mapStatus(richStatus()).GetRetry()

	if got.GetAttempts() != 3 || got.GetDelayMs() != 200 || got.GetMaxDelayMs() != 2000 {
		t.Fatalf("policy = %+v", got)
	}
	if got.GetTotalRetries() != 21 || got.GetExhaustedPolls() != 22 {
		t.Fatalf("counters = %d/%d", got.GetTotalRetries(), got.GetExhaustedPolls())
	}
}

func TestMapStatusCarriesTheClock(t *testing.T) {
	got := mapStatus(richStatus()).GetClock()

	if got.GetState() != "warn" {
		t.Fatalf("State = %q", got.GetState())
	}
	// Skew is signed, and a mapper that took an absolute value would hide the
	// difference between a device that runs fast and one that runs slow.
	if got.GetSkewMs() != -2500 || got.GetMedianSkewMs() != -2400 {
		t.Fatalf("skew = %d, median = %d", got.GetSkewMs(), got.GetMedianSkewMs())
	}
	if got.GetMinSkewMs() != -3000 || got.GetMaxSkewMs() != -1800 {
		t.Fatalf("range = %d..%d", got.GetMinSkewMs(), got.GetMaxSkewMs())
	}
	if got.GetDriftPerDayMs() != 24_000 || !got.GetDriftDetermined() || got.GetDriftFit() != 0.94 {
		t.Fatalf("drift = %+v", got)
	}
	if got.GetSamples() != 31 || got.GetObservedSamples() != 32 || got.GetRejectedSamples() != 33 {
		t.Fatalf("samples = %d/%d/%d", got.GetSamples(), got.GetObservedSamples(), got.GetRejectedSamples())
	}
	if got.GetWarnThresholdMs() != 2000 || got.GetCriticalThresholdMs() != 30_000 {
		t.Fatalf("thresholds = %d/%d", got.GetWarnThresholdMs(), got.GetCriticalThresholdMs())
	}
}

func TestMapStatusCarriesHealthAndIdentity(t *testing.T) {
	status := mapStatus(richStatus())
	health := status.GetHealth()
	identity := status.GetIdentity()

	if health.GetState() != "degraded" || health.GetAvailability() != 0.87 {
		t.Fatalf("health = %+v", health)
	}
	if health.GetLatencyP50Ms() != 44 || health.GetLatencyP95Ms() != 45 || health.GetLatencyP99Ms() != 46 {
		t.Fatalf("percentiles = %d/%d/%d", health.GetLatencyP50Ms(), health.GetLatencyP95Ms(), health.GetLatencyP99Ms())
	}
	if health.GetDegradeAfter() != 3 || health.GetOfflineAfter() != 10 || health.GetRecoverAfter() != 2 {
		t.Fatalf("hysteresis = %+v", health)
	}
	if !identity.GetKnown() || identity.GetModel() != "TTR20" || identity.GetSerial() != "SN-42" {
		t.Fatalf("identity = %+v", identity)
	}
}

func TestMapStatusLeavesUnsetTimestampsNil(t *testing.T) {
	got := mapStatus(domain.Status{})

	if got.GetLastSuccessfulReadTime() != nil || got.GetLastDeviceTime() != nil {
		t.Fatal("a gateway that has never read must not report a read time")
	}
	if got.GetClock().GetUpdatedAt() != nil || got.GetIdentity().GetReadAt() != nil {
		t.Fatal("unset nested timestamps must stay nil")
	}
	// The nested messages themselves are always present, so a consumer can
	// read through them without a nil check.
	if got.GetSchedule() == nil || got.GetRetry() == nil || got.GetClock() == nil {
		t.Fatal("nested sections must be present even when empty")
	}
}

func TestMapFleetCountsAndOrders(t *testing.T) {
	statuses := []domain.Status{
		{DeviceID: "alpha", Running: true, Health: domain.HealthStatus{State: "online"}, Clock: domain.ClockStatus{State: "ok", Skew: time.Second}},
		{DeviceID: "beta", Running: true, Health: domain.HealthStatus{State: "degraded"}, Clock: domain.ClockStatus{State: "warn", Skew: -4 * time.Second}},
		{DeviceID: "gamma", Health: domain.HealthStatus{State: "offline"}, Clock: domain.ClockStatus{State: "critical", Skew: 2 * time.Second}},
	}

	got := mapFleet(domain.SummarizeFleet(statuses))
	summary := got.GetSummary()

	if summary.GetDevices() != 3 || summary.GetRunning() != 2 {
		t.Fatalf("devices/running = %d/%d", summary.GetDevices(), summary.GetRunning())
	}
	if summary.GetOnline() != 1 || summary.GetDegraded() != 1 || summary.GetOffline() != 1 {
		t.Fatalf("health counts = %+v", summary)
	}
	if summary.GetClockOk() != 1 || summary.GetClockWarn() != 1 || summary.GetClockCritical() != 1 {
		t.Fatalf("clock counts = %+v", summary)
	}
	// "Worst" is by magnitude, so the device 4s behind beats the one 2s ahead.
	if summary.GetWorstClockDeviceId() != "beta" || summary.GetWorstClockSkewMs() != -4000 {
		t.Fatalf("worst = %q %d", summary.GetWorstClockDeviceId(), summary.GetWorstClockSkewMs())
	}
	if len(got.GetDevices()) != 3 {
		t.Fatalf("devices = %d", len(got.GetDevices()))
	}
}

func TestMapFleetOfNothing(t *testing.T) {
	got := mapFleet(domain.SummarizeFleet(nil))

	if got.GetSummary().GetDevices() != 0 {
		t.Fatalf("devices = %d", got.GetSummary().GetDevices())
	}
	if got.GetSummary().GetWorstClockDeviceId() != "" {
		t.Fatal("an empty fleet has no worst device")
	}
}

func TestGetFleetWithoutASupervisorReportsOneDevice(t *testing.T) {
	service := newTestGatewayService(t)

	got, err := New(context.Background(), service).GetFleet(context.Background(), &ft12v1.GetFleetRequest{})
	if err != nil {
		t.Fatal(err)
	}
	// A gateway without a device inventory still answers the fleet question,
	// so the console renders one table in both configurations.
	if got.GetSummary().GetDevices() != 1 || len(got.GetDevices()) != 1 {
		t.Fatalf("fleet = %+v", got.GetSummary())
	}
	if got.GetDevices()[0].GetTargetAddr() != "127.0.0.1:9000" {
		t.Fatalf("target = %q", got.GetDevices()[0].GetTargetAddr())
	}
}

type stubFleet struct{ fleet domain.FleetStatus }

func (s stubFleet) Fleet() domain.FleetStatus { return s.fleet }

func TestGetFleetPrefersTheSupervisor(t *testing.T) {
	service := newTestGatewayService(t)
	fleet := domain.SummarizeFleet([]domain.Status{
		{DeviceID: "alpha", Running: true, Health: domain.HealthStatus{State: "online"}},
		{DeviceID: "beta", Health: domain.HealthStatus{State: "offline"}},
	})

	api := New(context.Background(), service).WithFleet(stubFleet{fleet: fleet})
	got, err := api.GetFleet(context.Background(), &ft12v1.GetFleetRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if got.GetSummary().GetDevices() != 2 {
		t.Fatalf("devices = %d, want the supervisor view", got.GetSummary().GetDevices())
	}
}

func newTestGatewayService(t *testing.T) *domain.Service {
	t.Helper()
	service, err := domain.NewService(&config.GatewayConfig{
		Target:         "127.0.0.1:9000",
		CRCMode:        "sum",
		AdapterAddr:    1,
		PollInterval:   time.Second,
		RequestTimeout: time.Second,
		ConnectTimeout: time.Second,
		BackoffInitial: 10 * time.Millisecond,
		BackoffMax:     50 * time.Millisecond,
		RecentSize:     10,
	}, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	return service
}
