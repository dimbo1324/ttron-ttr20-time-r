package frame

import (
	"bytes"
	"errors"
	"testing"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
)

func buildFrame(t *testing.T, mode checksum.Mode, control, address byte, data []byte) []byte {
	t.Helper()
	raw, err := Encode(New(control, address, data), mode)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestBuildSkeletonAndAppendChecksum(t *testing.T) {
	tests := []struct {
		name string
		mode string
	}{
		{name: "sum", mode: string(checksum.ModeSum)},
		{name: "crc16", mode: string(checksum.ModeCRC16)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skeleton := BuildSkeleton(0x00, 0x01, []byte{0x01})
			raw := AppendChecksum(skeleton, tt.mode)

			mode, err := checksum.ParseMode(tt.mode)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := Decode(raw, mode)
			if err != nil {
				t.Fatalf("Decode() = %v", err)
			}
			if decoded.Control != 0x00 || decoded.Address != 0x01 {
				t.Fatalf("decoded = %+v", decoded)
			}
			if got := decoded.DataBytes(); len(got) != 1 || got[0] != 0x01 {
				t.Fatalf("payload = % X", got)
			}
		})
	}
}

func TestAppendChecksumRejectsBadInput(t *testing.T) {
	tests := []struct {
		name string
		raw  []byte
		mode string
	}{
		{name: "unknown mode", raw: []byte{StartByte, 0x03, StartByte, 0x00, 0x01, 0x01}, mode: "md5"},
		{name: "too short", raw: []byte{StartByte, 0x03}, mode: string(checksum.ModeSum)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AppendChecksum(tt.raw, tt.mode)
			if !bytes.Equal(got, tt.raw) {
				t.Fatalf("AppendChecksum() = % X, want the input unchanged", got)
			}
		})
	}
}

func TestVerify(t *testing.T) {
	sumFrame := buildFrame(t, checksum.ModeSum, 0x00, 0x01, []byte{0x01})
	crcFrame := buildFrame(t, checksum.ModeCRC16, 0x00, 0x01, []byte{0x01})

	if err := Verify(sumFrame); err != nil {
		t.Fatalf("Verify(sum) = %v", err)
	}
	if err := Verify(crcFrame); err != nil {
		t.Fatalf("Verify(crc16) = %v", err)
	}

	corrupt := append([]byte(nil), sumFrame...)
	corrupt[len(corrupt)-2] ^= 0xFF
	if err := Verify(corrupt); !errors.Is(err, ErrInvalidChecksum) {
		t.Fatalf("Verify(corrupt) = %v, want %v", err, ErrInvalidChecksum)
	}
	if err := Verify(nil); !errors.Is(err, ErrInvalidChecksum) {
		t.Fatalf("Verify(nil) = %v, want %v", err, ErrInvalidChecksum)
	}
}

func TestExtractFrame(t *testing.T) {
	tests := []struct {
		name string
		mode checksum.Mode
	}{
		{name: "sum", mode: checksum.ModeSum},
		{name: "crc16", mode: checksum.ModeCRC16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := buildFrame(t, tt.mode, 0x00, 0x01, []byte{0x01})
			buffer := bytes.NewBuffer(append(append([]byte(nil), raw...), 0xAA, 0xBB))

			got, ok := ExtractFrame(buffer)
			if !ok {
				t.Fatal("ExtractFrame() = false, want a frame")
			}
			if !bytes.Equal(got, raw) {
				t.Fatalf("ExtractFrame() = % X, want % X", got, raw)
			}
			if remaining := buffer.Bytes(); !bytes.Equal(remaining, []byte{0xAA, 0xBB}) {
				t.Fatalf("buffer remainder = % X, want the trailing bytes", remaining)
			}
		})
	}
}

func TestExtractFrameWithoutCompleteFrame(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{StartByte, 0x03, StartByte, 0x00})

	if _, ok := ExtractFrame(buffer); ok {
		t.Fatal("ExtractFrame() = true, want false for a partial frame")
	}
	if buffer.Len() != 4 {
		t.Fatalf("buffer length = %d, want the input preserved", buffer.Len())
	}
}

func TestCorruptChecksum(t *testing.T) {
	tests := []struct {
		name string
		mode checksum.Mode
	}{
		{name: "sum", mode: checksum.ModeSum},
		{name: "crc16", mode: checksum.ModeCRC16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := buildFrame(t, tt.mode, 0x00, 0x01, []byte{0x01})
			corrupt := append([]byte(nil), raw...)

			CorruptChecksum(corrupt, string(tt.mode))
			if bytes.Equal(corrupt, raw) {
				t.Fatal("CorruptChecksum() left the frame untouched")
			}
			if _, err := Decode(corrupt, tt.mode); !errors.Is(err, ErrInvalidChecksum) {
				t.Fatalf("Decode() = %v, want %v", err, ErrInvalidChecksum)
			}
			if corrupt[len(corrupt)-1] != EndByte {
				t.Fatal("CorruptChecksum() must keep the end byte intact")
			}
		})
	}
}

func TestCorruptChecksumIgnoresUnusableInput(t *testing.T) {
	short := []byte{StartByte}
	original := append([]byte(nil), short...)
	CorruptChecksum(short, string(checksum.ModeSum))
	if !bytes.Equal(short, original) {
		t.Fatalf("CorruptChecksum() = % X, want the input unchanged", short)
	}

	raw := buildFrame(t, checksum.ModeSum, 0x00, 0x01, []byte{0x01})
	corrupt := append([]byte(nil), raw...)
	CorruptChecksum(corrupt, "md5")
	if bytes.Equal(corrupt, raw) {
		t.Fatal("an unknown mode must fall back to the sum layout")
	}
}

func TestPayloadData(t *testing.T) {
	raw := buildFrame(t, checksum.ModeSum, 0x00, 0x01, []byte{0x01, 0x02})

	got := PayloadData(raw)
	want := []byte{0x01, 0x02}
	if !bytes.Equal(got, want) {
		t.Fatalf("PayloadData() = % X, want the command data % X", got, want)
	}

	crcFrame := buildFrame(t, checksum.ModeCRC16, 0x00, 0x01, []byte{0x01, 0x02})
	if got := PayloadData(crcFrame); !bytes.Equal(got, want) {
		t.Fatalf("PayloadData(crc16) = % X, want % X", got, want)
	}
}

func TestPayloadDataRejectsMalformedFrames(t *testing.T) {
	tests := []struct {
		name string
		raw  []byte
	}{
		{name: "too short", raw: []byte{StartByte, 0x03, StartByte}},
		{name: "payload length below minimum", raw: []byte{StartByte, 0x01, StartByte, 0x00, 0x01, EndByte}},
		{name: "payload length beyond frame", raw: []byte{StartByte, 0x40, StartByte, 0x00, 0x01, 0x01, 0x02, EndByte}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PayloadData(tt.raw); got != nil {
				t.Fatalf("PayloadData() = % X, want nil", got)
			}
		})
	}
}

func TestCRC16CompatibilityAlias(t *testing.T) {
	if CRC16 != string(checksum.ModeCRC16) {
		t.Fatalf("CRC16 = %q, want %q", CRC16, checksum.ModeCRC16)
	}
}
