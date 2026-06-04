package checksum

import (
	"errors"
	"testing"
)

func TestSum8(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want byte
	}{
		{name: "empty", data: nil, want: 0x00},
		{name: "read-time request body", data: []byte{0x00, 0x01, 0x01}, want: 0x02},
		{name: "wrap", data: []byte{0xFF, 0x02}, want: 0x01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sum8(tt.data); got != tt.want {
				t.Fatalf("Sum8() = 0x%02X, want 0x%02X", got, tt.want)
			}
		})
	}
}

func TestCRC16(t *testing.T) {
	if got := CRC16([]byte("123456789")); got != 0x4B37 {
		t.Fatalf("CRC16() = 0x%04X, want 0x4B37", got)
	}
}

func TestCRC16AdditionalVectors(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want uint16
	}{
		{name: "empty", data: nil, want: 0xFFFF},
		{name: "single zero", data: []byte{0x00}, want: 0x40BF},
		{name: "read-time payload", data: []byte{0x00, 0x01, 0x01}, want: 0x90B1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CRC16(tt.data); got != tt.want {
				t.Fatalf("CRC16() = 0x%04X, want 0x%04X", got, tt.want)
			}
		})
	}
}

func TestParseMode(t *testing.T) {
	tests := []struct {
		raw     string
		want    Mode
		wantErr bool
	}{
		{raw: "", want: ModeSum},
		{raw: " sum ", want: ModeSum},
		{raw: "sum", want: ModeSum},
		{raw: "SUM", want: ModeSum},
		{raw: "crc16", want: ModeCRC16},
		{raw: "CRC16", want: ModeCRC16},
		{raw: "bad", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			got, err := ParseMode(tt.raw)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidMode) {
					t.Fatalf("ParseMode() error = %v, want ErrInvalidMode", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseMode() unexpected error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ParseMode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestComputeAndVerify(t *testing.T) {
	tests := []struct {
		name string
		mode Mode
		data []byte
		want []byte
	}{
		{name: "sum", mode: ModeSum, data: []byte{0x00, 0x01, 0x01}, want: []byte{0x02}},
		{name: "crc16", mode: ModeCRC16, data: []byte("123456789"), want: []byte{0x37, 0x4B}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Compute(tt.mode, tt.data)
			if err != nil {
				t.Fatalf("Compute() error = %v", err)
			}
			if string(got) != string(tt.want) {
				t.Fatalf("Compute() = % X, want % X", got, tt.want)
			}
			if err := Verify(tt.mode, tt.data, tt.want); err != nil {
				t.Fatalf("Verify() error = %v", err)
			}
			bad := append([]byte(nil), tt.want...)
			bad[0] ^= 0xFF
			if err := Verify(tt.mode, tt.data, bad); !errors.Is(err, ErrInvalidChecksum) {
				t.Fatalf("Verify() error = %v, want ErrInvalidChecksum", err)
			}
		})
	}
}

func TestChecksumLength(t *testing.T) {
	tests := []struct {
		mode Mode
		want int
	}{
		{mode: ModeSum, want: 1},
		{mode: ModeCRC16, want: 2},
	}
	for _, tt := range tests {
		got, err := tt.mode.ChecksumLength()
		if err != nil {
			t.Fatalf("ChecksumLength() error = %v", err)
		}
		if got != tt.want {
			t.Fatalf("ChecksumLength() = %d, want %d", got, tt.want)
		}
	}
}
