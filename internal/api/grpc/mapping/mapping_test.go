package mapping

import (
	"testing"
	"time"

	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/observability/events"
)

func TestChecksumMode(t *testing.T) {
	if ChecksumMode("sum") != ft12v1.ChecksumMode_CHECKSUM_MODE_SUM {
		t.Fatal("sum did not map to CHECKSUM_MODE_SUM")
	}
	if ChecksumMode("crc16") != ft12v1.ChecksumMode_CHECKSUM_MODE_CRC16 {
		t.Fatal("crc16 did not map to CHECKSUM_MODE_CRC16")
	}
	if ChecksumMode("bad") != ft12v1.ChecksumMode_CHECKSUM_MODE_UNSPECIFIED {
		t.Fatal("bad checksum did not map to unspecified")
	}
}

func TestDirection(t *testing.T) {
	tests := map[events.Direction]ft12v1.EventDirection{
		events.DirectionRX:     ft12v1.EventDirection_EVENT_DIRECTION_RX,
		events.DirectionTX:     ft12v1.EventDirection_EVENT_DIRECTION_TX,
		events.DirectionError:  ft12v1.EventDirection_EVENT_DIRECTION_ERROR,
		events.DirectionDrop:   ft12v1.EventDirection_EVENT_DIRECTION_SYSTEM,
		events.DirectionSystem: ft12v1.EventDirection_EVENT_DIRECTION_SYSTEM,
	}
	for input, want := range tests {
		if got := Direction(input); got != want {
			t.Fatalf("Direction(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestDirectionUnknownMapsToSystem(t *testing.T) {
	if got := Direction(events.Direction("unknown")); got != ft12v1.EventDirection_EVENT_DIRECTION_SYSTEM {
		t.Fatalf("Direction(unknown) = %v", got)
	}
}

func TestEventsKeepsStableIDAndLimit(t *testing.T) {
	ts := time.Unix(10, 0).UTC()
	records := []events.FrameRecord{
		{ID: 11, Timestamp: ts, Direction: events.DirectionRX, Service: events.ServiceGateway, ChecksumMode: "sum", RawHex: "one"},
		{ID: 12, Timestamp: ts, Direction: events.DirectionTX, Service: events.ServiceGateway, ChecksumMode: "crc16", RawHex: "two"},
	}

	got := Events(records, 1)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].GetId() != 12 || got[0].GetRawHex() != "two" {
		t.Fatalf("event = %+v, want id 12/raw two", got[0])
	}
	if got[0].GetTimestamp() == nil {
		t.Fatal("timestamp is nil, want mapped timestamp")
	}
}

func TestTimeZeroReturnsNil(t *testing.T) {
	if Time(time.Time{}) != nil {
		t.Fatal("zero time should map to nil")
	}
}

func TestServiceState(t *testing.T) {
	tests := []struct {
		name      string
		running   bool
		lastError string
		want      ft12v1.ServiceState
	}{
		{name: "stopped", running: false, want: ft12v1.ServiceState_SERVICE_STATE_STOPPED},
		{name: "running", running: true, want: ft12v1.ServiceState_SERVICE_STATE_RUNNING},
		{name: "degraded", running: true, lastError: "timeout", want: ft12v1.ServiceState_SERVICE_STATE_DEGRADED},
		{name: "stopped with stale error", running: false, lastError: "timeout", want: ft12v1.ServiceState_SERVICE_STATE_STOPPED},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ServiceState(tt.running, tt.lastError); got != tt.want {
				t.Fatalf("ServiceState(%v, %q) = %v, want %v", tt.running, tt.lastError, got, tt.want)
			}
		})
	}
}
