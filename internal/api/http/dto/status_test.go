package dto

import (
	"testing"
	"time"

	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestTimestampUsesUTCAndNilSafe(t *testing.T) {
	if Timestamp(nil) != nil {
		t.Fatal("nil timestamp should map to nil")
	}
	input := timestamppb.New(time.Date(2026, 6, 4, 12, 30, 0, 123, time.FixedZone("MSK", 3*60*60)))
	got := Timestamp(input)
	if got == nil {
		t.Fatal("timestamp should not be nil")
	}
	if *got != "2026-06-04T09:30:00.000000123Z" {
		t.Fatalf("timestamp = %q", *got)
	}
}

func TestStatusAndProtocolValueMappings(t *testing.T) {
	stateCases := map[ft12v1.ServiceState]string{
		ft12v1.ServiceState_SERVICE_STATE_STOPPED:     "stopped",
		ft12v1.ServiceState_SERVICE_STATE_RUNNING:     "running",
		ft12v1.ServiceState_SERVICE_STATE_DEGRADED:    "degraded",
		ft12v1.ServiceState_SERVICE_STATE_UNSPECIFIED: "unspecified",
	}
	for input, want := range stateCases {
		if got := ServiceState(input); got != want {
			t.Fatalf("ServiceState(%v) = %q, want %q", input, got, want)
		}
	}

	if ChecksumMode(ft12v1.ChecksumMode_CHECKSUM_MODE_SUM) != "sum" {
		t.Fatal("sum checksum did not map")
	}
	if ChecksumMode(ft12v1.ChecksumMode_CHECKSUM_MODE_CRC16) != "crc16" {
		t.Fatal("crc16 checksum did not map")
	}
	if Direction(ft12v1.EventDirection_EVENT_DIRECTION_ERROR) != "ERR" {
		t.Fatal("error direction did not map to ERR")
	}
	if EventDirectionFromString("error") != ft12v1.EventDirection_EVENT_DIRECTION_ERROR {
		t.Fatal("error string did not map to ERROR direction")
	}
	if EventDirectionFromString("unknown") != ft12v1.EventDirection_EVENT_DIRECTION_SYSTEM {
		t.Fatal("unknown string should map to SYSTEM")
	}
}

func TestNilStatusDTOsAreFrontendSafe(t *testing.T) {
	emu := EmulatorStatus(nil)
	if emu.State != "unavailable" || emu.ChecksumMode != "unspecified" {
		t.Fatalf("emulator nil status = %+v", emu)
	}
	gateway := GatewayStatus(nil)
	if gateway.State != "unavailable" || gateway.ChecksumMode != "unspecified" {
		t.Fatalf("gateway nil status = %+v", gateway)
	}
	if FaultMode(nil) != nil {
		t.Fatal("nil fault mode should stay nil")
	}
	if LastReadTime(nil).Available {
		t.Fatal("nil last-read response should not be available")
	}
}
