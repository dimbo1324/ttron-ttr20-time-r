package frame

import (
	"bytes"
	"errors"
	"testing"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
)

func TestEncodeDecodeRoundtrip(t *testing.T) {
	tests := []struct {
		name string
		mode checksum.Mode
		want []byte
	}{
		{
			name: "sum",
			mode: checksum.ModeSum,
			want: []byte{0x68, 0x03, 0x68, 0x00, 0x01, 0x01, 0x02, 0x16},
		},
		{
			name: "crc16",
			mode: checksum.ModeCRC16,
			want: []byte{0x68, 0x03, 0x68, 0x00, 0x01, 0x01, 0xB1, 0x90, 0x16},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, err := Encode(New(0x00, 0x01, []byte{0x01}), tt.mode)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if !bytes.Equal(raw, tt.want) {
				t.Fatalf("Encode() = % X, want % X", raw, tt.want)
			}

			got, err := Decode(raw, tt.mode)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if got.Control != 0x00 || got.Address != 0x01 || !bytes.Equal(got.DataBytes(), []byte{0x01}) {
				t.Fatalf("Decode() = %+v", got)
			}
			if !bytes.Equal(got.RawBytes(), raw) {
				t.Fatalf("RawBytes() = % X, want % X", got.RawBytes(), raw)
			}
		})
	}
}

func TestDecodeErrors(t *testing.T) {
	valid, err := Encode(New(0x00, 0x01, []byte{0x01}), checksum.ModeSum)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		raw  []byte
		want error
	}{
		{name: "too short", raw: []byte{0x68}, want: ErrFrameTooShort},
		{name: "invalid first start", raw: mutate(valid, 0, 0x67), want: ErrInvalidStartByte},
		{name: "invalid second start", raw: mutate(valid, 2, 0x67), want: ErrInvalidRepeatedStartByte},
		{name: "invalid length too small", raw: mutate(valid, 1, 0x01), want: ErrInvalidLength},
		{name: "invalid len mismatch longer than data", raw: mutate(valid, 1, 0x04), want: ErrInvalidLength},
		{name: "invalid checksum", raw: mutate(valid, len(valid)-2, 0x99), want: ErrInvalidChecksum},
		{name: "truncated checksum", raw: valid[:len(valid)-2], want: ErrFrameTooShort},
		{name: "invalid end", raw: mutate(valid, len(valid)-1, 0x00), want: ErrInvalidEndByte},
		{name: "invalid byte count", raw: valid[:len(valid)-1], want: ErrInvalidLength},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decode(tt.raw, checksum.ModeSum)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Decode() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestEncodeRejectsOversizedPayload(t *testing.T) {
	data := make([]byte, 254)
	_, err := Encode(New(0x00, 0x01, data), checksum.ModeSum)
	if !errors.Is(err, ErrUnsupportedPayloadLength) {
		t.Fatalf("Encode() error = %v, want ErrUnsupportedPayloadLength", err)
	}
}

func TestEmptyDataFrame(t *testing.T) {
	raw, err := Encode(New(0x10, 0x02, nil), checksum.ModeSum)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	got, err := Decode(raw, checksum.ModeSum)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if got.Control != 0x10 || got.Address != 0x02 || len(got.DataBytes()) != 0 {
		t.Fatalf("Decode() = %+v", got)
	}
}

func TestCompatibilityHelpers(t *testing.T) {
	raw := AppendChecksum(BuildSkeleton(0x00, 0x01, []byte{0x01}), "sum")
	if err := Verify(raw); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	var buf bytes.Buffer
	buf.Write(raw)
	got, ok := ExtractFrame(&buf)
	if !ok {
		t.Fatal("ExtractFrame() did not find a frame")
	}
	if !bytes.Equal(got, raw) {
		t.Fatalf("ExtractFrame() = % X, want % X", got, raw)
	}
}

func mutate(src []byte, index int, value byte) []byte {
	out := append([]byte(nil), src...)
	out[index] = value
	return out
}
