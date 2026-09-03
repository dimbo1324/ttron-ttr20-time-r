package dto

import (
	"encoding/json"
	"testing"
	"time"

	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGatewayStatusCarriesTheNestedSections(t *testing.T) {
	at := timestamppb.New(time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC))
	got := GatewayStatus(&ft12v1.GatewayStatus{
		State:          ft12v1.ServiceState_SERVICE_STATE_RUNNING,
		ChecksumMode:   ft12v1.ChecksumMode_CHECKSUM_MODE_CRC16,
		ProtocolErrors: 7,
		DeviceId:       "alpha",
		DeviceName:     "Alpha meter",
		Schedule:       &ft12v1.ScheduleStatus{Mode: "aligned", IntervalMs: 60_000, OffsetMs: 5000, NextPollAt: at},
		Retry:          &ft12v1.RetryStatus{Attempts: 3, TotalRetries: 9},
		Clock:          &ft12v1.ClockStatus{State: "warn", SkewMs: -2500, DriftPerDayMs: 24_000, DriftFit: 0.9},
		Health:         &ft12v1.DeviceHealth{State: "degraded", Availability: 0.75, LatencyP95Ms: 45},
		Identity:       &ft12v1.DeviceIdentity{Known: true, Model: "TTR20", ReadAt: at},
	})

	if got.State != "running" || got.ChecksumMode != "crc16" {
		t.Fatalf("state/mode = %q/%q", got.State, got.ChecksumMode)
	}
	if got.ProtocolErrors != 7 || got.DeviceID != "alpha" {
		t.Fatalf("status = %+v", got)
	}
	if got.Schedule.Mode != "aligned" || got.Schedule.NextPollAt == nil {
		t.Fatalf("schedule = %+v", got.Schedule)
	}
	if got.Retry.Attempts != 3 || got.Retry.TotalRetries != 9 {
		t.Fatalf("retry = %+v", got.Retry)
	}
	if got.Clock.State != "warn" || got.Clock.SkewMs != -2500 {
		t.Fatalf("clock = %+v", got.Clock)
	}
	if got.Health.State != "degraded" || got.Health.LatencyP95Ms != 45 {
		t.Fatalf("health = %+v", got.Health)
	}
	if !got.Identity.Known || got.Identity.Model != "TTR20" || got.Identity.ReadAt == nil {
		t.Fatalf("identity = %+v", got.Identity)
	}
}

func TestGatewayStatusNamesTheUnknownStates(t *testing.T) {
	// The console keys colour and wording off these strings; an empty one is
	// not a state anything can name, so the DTO fills it in.
	got := GatewayStatus(&ft12v1.GatewayStatus{})

	if got.Clock.State != "unknown" || got.Health.State != "unknown" {
		t.Fatalf("states = %q/%q", got.Clock.State, got.Health.State)
	}
}

func TestGatewayStatusOfNothingIsStillWellShaped(t *testing.T) {
	got := GatewayStatus(nil)

	if got.State != "unavailable" || got.ChecksumMode != "unspecified" {
		t.Fatalf("status = %+v", got)
	}
	if got.Clock.State != "unknown" || got.Health.State != "unknown" {
		t.Fatalf("states = %q/%q", got.Clock.State, got.Health.State)
	}
}

func TestFleetMapsSummaryAndDevices(t *testing.T) {
	got := Fleet(&ft12v1.GetFleetResponse{
		Summary: &ft12v1.FleetSummary{
			Devices: 3, Running: 2, Online: 1, Degraded: 1, Offline: 1,
			ClockOk: 1, ClockWarn: 1, ClockCritical: 1,
			WorstClockSkewMs: -4000, WorstClockDeviceId: "beta",
		},
		Devices: []*ft12v1.GatewayStatus{{DeviceId: "alpha"}, {DeviceId: "beta"}, {DeviceId: "gamma"}},
	})

	if got.Summary.Devices != 3 || got.Summary.Running != 2 {
		t.Fatalf("summary = %+v", got.Summary)
	}
	if got.Summary.WorstClockDeviceID != "beta" || got.Summary.WorstClockSkewMs != -4000 {
		t.Fatalf("worst = %+v", got.Summary)
	}
	if len(got.Devices) != 3 || got.Devices[1].DeviceID != "beta" {
		t.Fatalf("devices = %+v", got.Devices)
	}
}

func TestFleetOfNothingSerialisesAsAnEmptyList(t *testing.T) {
	// A null here would make every consumer guard before iterating; an empty
	// array reads the same as a fleet that happens to have no devices.
	body, err := json.Marshal(Fleet(nil))
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Devices []GatewayStatusDTO `json:"devices"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Devices == nil {
		t.Fatalf("devices serialised as null: %s", body)
	}
	if len(decoded.Devices) != 0 {
		t.Fatalf("devices = %d", len(decoded.Devices))
	}
}
