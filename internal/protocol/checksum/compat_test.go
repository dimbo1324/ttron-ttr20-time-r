package checksum

import (
	"errors"
	"testing"
)

func TestSumAlias(t *testing.T) {
	data := []byte{0x00, 0x01, 0x01}
	if got, want := Sum(data), Sum8(data); got != want {
		t.Fatalf("Sum() = 0x%02X, want 0x%02X", got, want)
	}
}

func TestCRC16ModbusAlias(t *testing.T) {
	data := []byte("123456789")
	if got, want := CRC16Modbus(data), CRC16(data); got != want {
		t.Fatalf("CRC16Modbus() = 0x%04X, want 0x%04X", got, want)
	}
}

func TestMustParseMode(t *testing.T) {
	if got := MustParseMode("crc16"); got != ModeCRC16 {
		t.Fatalf("MustParseMode() = %q, want %q", got, ModeCRC16)
	}

	defer func() {
		if recover() == nil {
			t.Fatal("MustParseMode() must panic on an unknown mode")
		}
	}()
	MustParseMode("md5")
}

func TestChecksumLengthRejectsUnknownMode(t *testing.T) {
	if _, err := Mode("md5").ChecksumLength(); !errors.Is(err, ErrInvalidMode) {
		t.Fatalf("ChecksumLength() error = %v, want %v", err, ErrInvalidMode)
	}

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
			t.Fatalf("ChecksumLength(%q) = %d, want %d", tt.mode, got, tt.want)
		}
	}
}

func TestComputeAndVerifyRejectUnknownMode(t *testing.T) {
	if _, err := Compute(Mode("md5"), []byte{0x01}); !errors.Is(err, ErrInvalidMode) {
		t.Fatalf("Compute() error = %v, want %v", err, ErrInvalidMode)
	}
	if err := Verify(Mode("md5"), []byte{0x01}, []byte{0x01}); !errors.Is(err, ErrInvalidMode) {
		t.Fatalf("Verify() error = %v, want %v", err, ErrInvalidMode)
	}
}

func TestComputeAndVerifyRoundTrip(t *testing.T) {
	payload := []byte{0x00, 0x01, 0x01}

	for _, mode := range []Mode{ModeSum, ModeCRC16} {
		t.Run(string(mode), func(t *testing.T) {
			sum, err := Compute(mode, payload)
			if err != nil {
				t.Fatal(err)
			}
			length, err := mode.ChecksumLength()
			if err != nil {
				t.Fatal(err)
			}
			if len(sum) != length {
				t.Fatalf("Compute() returned %d bytes, want %d", len(sum), length)
			}
			if err := Verify(mode, payload, sum); err != nil {
				t.Fatalf("Verify() = %v", err)
			}

			corrupt := append([]byte(nil), sum...)
			corrupt[0] ^= 0xFF
			if err := Verify(mode, payload, corrupt); err == nil {
				t.Fatal("Verify() must reject a corrupted checksum")
			}
		})
	}
}

func TestVerifyRejectsWrongChecksumLength(t *testing.T) {
	if err := Verify(ModeCRC16, []byte{0x01}, []byte{0x01}); err == nil {
		t.Fatal("Verify() must reject a truncated crc16 checksum")
	}
}
