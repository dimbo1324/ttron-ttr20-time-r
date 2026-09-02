package emulator

import (
	"testing"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/codec"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/command"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
)

func TestBuildTimeResponse(t *testing.T) {
	tests := []struct {
		name string
		mode string
		want checksum.Mode
	}{
		{name: "sum", mode: "sum", want: checksum.ModeSum},
		{name: "crc16", mode: "crc16", want: checksum.ModeCRC16},
		{name: "unknown mode falls back to sum", mode: "md5", want: checksum.ModeSum},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := BuildTimeResponse(0x00, 0x01, command.BuildReadTimeRequest(), tt.mode, 0)
			if len(raw) == 0 {
				t.Fatal("BuildTimeResponse() returned no frame")
			}

			_, parsed, err := codec.New(tt.want, 0x00, 0x01).DecodeReadTimeResponse(raw)
			if err != nil {
				t.Fatalf("DecodeReadTimeResponse() = %v", err)
			}
			if delta := time.Since(parsed.Time); delta > 5*time.Second || delta < -5*time.Second {
				t.Fatalf("device time = %s, want about now", parsed.Time)
			}
		})
	}
}

func TestBuildAckResponse(t *testing.T) {
	tests := []struct {
		name string
		mode string
		want checksum.Mode
	}{
		{name: "sum", mode: "sum", want: checksum.ModeSum},
		{name: "crc16", mode: "crc16", want: checksum.ModeCRC16},
		{name: "unknown mode falls back to sum", mode: "md5", want: checksum.ModeSum},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := BuildAckResponse(0x00, 0x01, []byte{0x33}, tt.mode, 0)
			if len(raw) == 0 {
				t.Fatal("BuildAckResponse() returned no frame")
			}

			decoded, err := frame.Decode(raw, tt.want)
			if err != nil {
				t.Fatalf("Decode() = %v", err)
			}
			payload := decoded.DataBytes()
			if payload[0] != 0x33 || string(payload[1:]) != "OK" {
				t.Fatalf("ack payload = % X", payload)
			}
			if decoded.Control != 0x80 {
				t.Fatalf("ack control = 0x%02X, want the response bit set", decoded.Control)
			}
		})
	}
}

func TestBuildAckResponseWithEmptyRequestData(t *testing.T) {
	raw := BuildAckResponse(0x00, 0x01, nil, "sum", 0)

	decoded, err := frame.Decode(raw, checksum.ModeSum)
	if err != nil {
		t.Fatal(err)
	}
	if got := decoded.DataBytes()[0]; got != 0xFF {
		t.Fatalf("ack command = 0x%02X, want the 0xFF fallback", got)
	}
}
