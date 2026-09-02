package dto

import (
	"testing"
	"time"

	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestChecksumMode(t *testing.T) {
	tests := []struct {
		name string
		in   ft12v1.ChecksumMode
		want string
	}{
		{name: "sum", in: ft12v1.ChecksumMode_CHECKSUM_MODE_SUM, want: "sum"},
		{name: "crc16", in: ft12v1.ChecksumMode_CHECKSUM_MODE_CRC16, want: "crc16"},
		{name: "unspecified", in: ft12v1.ChecksumMode_CHECKSUM_MODE_UNSPECIFIED, want: "unspecified"},
		{name: "unknown value", in: ft12v1.ChecksumMode(99), want: "unspecified"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ChecksumMode(tt.in); got != tt.want {
				t.Fatalf("ChecksumMode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDirection(t *testing.T) {
	tests := []struct {
		name string
		in   ft12v1.EventDirection
		want string
	}{
		{name: "rx", in: ft12v1.EventDirection_EVENT_DIRECTION_RX, want: "RX"},
		{name: "tx", in: ft12v1.EventDirection_EVENT_DIRECTION_TX, want: "TX"},
		{name: "error", in: ft12v1.EventDirection_EVENT_DIRECTION_ERROR, want: "ERR"},
		{name: "system", in: ft12v1.EventDirection_EVENT_DIRECTION_SYSTEM, want: "SYSTEM"},
		{name: "unknown value", in: ft12v1.EventDirection(99), want: "SYSTEM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Direction(tt.in); got != tt.want {
				t.Fatalf("Direction() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEventDirectionFromString(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want ft12v1.EventDirection
	}{
		{name: "rx", in: "rx", want: ft12v1.EventDirection_EVENT_DIRECTION_RX},
		{name: "tx upper", in: "TX", want: ft12v1.EventDirection_EVENT_DIRECTION_TX},
		{name: "err", in: "err", want: ft12v1.EventDirection_EVENT_DIRECTION_ERROR},
		{name: "error", in: "ERROR", want: ft12v1.EventDirection_EVENT_DIRECTION_ERROR},
		{name: "system", in: "system", want: ft12v1.EventDirection_EVENT_DIRECTION_SYSTEM},
		{name: "unknown", in: "whatever", want: ft12v1.EventDirection_EVENT_DIRECTION_SYSTEM},
		{name: "empty", in: "", want: ft12v1.EventDirection_EVENT_DIRECTION_SYSTEM},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EventDirectionFromString(tt.in); got != tt.want {
				t.Fatalf("EventDirectionFromString(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestDirectionRoundTrip(t *testing.T) {
	for _, direction := range []ft12v1.EventDirection{
		ft12v1.EventDirection_EVENT_DIRECTION_RX,
		ft12v1.EventDirection_EVENT_DIRECTION_TX,
		ft12v1.EventDirection_EVENT_DIRECTION_ERROR,
		ft12v1.EventDirection_EVENT_DIRECTION_SYSTEM,
	} {
		if got := EventDirectionFromString(Direction(direction)); got != direction {
			t.Fatalf("round trip of %v = %v", direction, got)
		}
	}
}

func TestEventMapping(t *testing.T) {
	moment := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	event := &ft12v1.FrameEvent{
		Id:           7,
		Timestamp:    timestamppb.New(moment),
		Service:      "gateway",
		Direction:    ft12v1.EventDirection_EVENT_DIRECTION_RX,
		RemoteAddr:   "127.0.0.1:9000",
		ChecksumMode: ft12v1.ChecksumMode_CHECKSUM_MODE_CRC16,
		RawHex:       "68 03 68",
		Command:      "read-time",
		Error:        "",
		Message:      "ok",
	}

	got := Event(event, "fallback")
	if got.ID != 7 || got.Source != "gateway" || got.Service != "gateway" {
		t.Fatalf("Event() = %+v", got)
	}
	if got.Direction != "RX" || got.ChecksumMode != "crc16" {
		t.Fatalf("Event() = %+v", got)
	}
	if got.RemoteAddr != "127.0.0.1:9000" || got.RawHex != "68 03 68" || got.Command != "read-time" {
		t.Fatalf("Event() = %+v", got)
	}
	if got.Message != "ok" {
		t.Fatalf("Event().Message = %q", got.Message)
	}
}

func TestEventUsesFallbackSource(t *testing.T) {
	event := &ft12v1.FrameEvent{Id: 1, Direction: ft12v1.EventDirection_EVENT_DIRECTION_TX}

	got := Event(event, "emulator")
	if got.Source != "emulator" || got.Service != "emulator" {
		t.Fatalf("Event() = %+v, want the fallback source", got)
	}
}

func TestEventHandlesNil(t *testing.T) {
	got := Event(nil, "emulator")

	if got.Source != "emulator" || got.Service != "emulator" {
		t.Fatalf("Event(nil) = %+v", got)
	}
	if got.Direction != "SYSTEM" || got.ChecksumMode != "unspecified" {
		t.Fatalf("Event(nil) = %+v", got)
	}
}

func TestEventsMapping(t *testing.T) {
	events := []*ft12v1.FrameEvent{
		{Id: 1, Service: "emulator", Direction: ft12v1.EventDirection_EVENT_DIRECTION_RX},
		{Id: 2, Direction: ft12v1.EventDirection_EVENT_DIRECTION_TX},
		nil,
	}

	got := Events(events, "gateway")
	if len(got) != 3 {
		t.Fatalf("Events() = %d entries, want 3", len(got))
	}
	if got[0].Source != "emulator" || got[1].Source != "gateway" || got[2].Source != "gateway" {
		t.Fatalf("Events() = %+v", got)
	}
}

func TestEventsWithEmptyInput(t *testing.T) {
	got := Events(nil, "gateway")
	if got == nil {
		t.Fatal("Events(nil) must return an empty slice, not nil, so JSON renders []")
	}
	if len(got) != 0 {
		t.Fatalf("Events(nil) = %d entries, want 0", len(got))
	}
}

func TestFaultModeProtoRoundTrip(t *testing.T) {
	request := FaultModeDTO{
		ResponseDelayMs:            250,
		CorruptChecksum:            true,
		CorruptChecksumProbability: 0.25,
		FragmentResponse:           true,
		FragmentProbability:        0.5,
		FragmentDelayMs:            40,
		NoResponse:                 true,
		CloseAfterRequest:          true,
	}

	proto := FaultModeProto(request)
	if proto.GetResponseDelayMs() != 250 || proto.GetFragmentDelayMs() != 40 {
		t.Fatalf("FaultModeProto() = %+v", proto)
	}
	if proto.GetCorruptChecksumProbability() != 0.25 || proto.GetFragmentProbability() != 0.5 {
		t.Fatalf("FaultModeProto() = %+v", proto)
	}
	if !proto.GetCorruptChecksum() || !proto.GetFragmentResponse() {
		t.Fatalf("FaultModeProto() = %+v", proto)
	}
	if !proto.GetNoResponse() || !proto.GetCloseAfterRequest() {
		t.Fatalf("FaultModeProto() = %+v", proto)
	}

	back := FaultMode(proto)
	if back == nil {
		t.Fatal("FaultMode() = nil, want a mapped value")
	}
	if *back != request {
		t.Fatalf("FaultMode() = %+v, want %+v", *back, request)
	}
}

func TestFaultModeHandlesNil(t *testing.T) {
	if got := FaultMode(nil); got != nil {
		t.Fatalf("FaultMode(nil) = %+v, want nil", got)
	}
}

func TestEmulatorStatusHandlesNil(t *testing.T) {
	got := EmulatorStatus(nil)
	if got.State == "" {
		t.Fatalf("EmulatorStatus(nil) = %+v, want a rendered state", got)
	}
}

func TestGatewayStatusHandlesNil(t *testing.T) {
	got := GatewayStatus(nil)
	if got.State == "" {
		t.Fatalf("GatewayStatus(nil) = %+v, want a rendered state", got)
	}
}

func TestLastReadTimeHandlesNil(t *testing.T) {
	got := LastReadTime(nil)
	if got.Available {
		t.Fatalf("LastReadTime(nil) = %+v, want unavailable", got)
	}
}
