package frame

import (
	"bytes"
	"errors"
	"testing"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
)

func TestStreamParserCompleteFrame(t *testing.T) {
	raw := mustEncode(t, checksum.ModeSum, []byte{0x01})
	parser := NewStreamParser(checksum.ModeSum)

	result := parser.Push(raw)
	if len(result.Errors) != 0 {
		t.Fatalf("Push() errors = %v", result.Errors)
	}
	if len(result.Frames) != 1 {
		t.Fatalf("Push() frames = %d, want 1", len(result.Frames))
	}
	if !bytes.Equal(result.Frames[0].RawBytes(), raw) {
		t.Fatalf("RawBytes() = % X, want % X", result.Frames[0].RawBytes(), raw)
	}
}

func TestStreamParserFragmentedFrame(t *testing.T) {
	raw := mustEncode(t, checksum.ModeSum, []byte{0x01})
	parser := NewStreamParser(checksum.ModeSum)

	result := parser.Push(raw[:3])
	if len(result.Frames) != 0 || parser.BufferedLen() == 0 {
		t.Fatalf("partial push result = %+v buffered=%d", result, parser.BufferedLen())
	}
	result = parser.Push(raw[3:])
	if len(result.Errors) != 0 || len(result.Frames) != 1 {
		t.Fatalf("final push result = %+v", result)
	}
}

func TestStreamParserByteByByte(t *testing.T) {
	raw := mustEncode(t, checksum.ModeSum, []byte{0x01})
	parser := NewStreamParser(checksum.ModeSum)
	var frames []Frame
	for _, b := range raw {
		result := parser.Push([]byte{b})
		frames = append(frames, result.Frames...)
		if len(result.Errors) != 0 {
			t.Fatalf("Push() errors = %v", result.Errors)
		}
	}
	if len(frames) != 1 {
		t.Fatalf("frames = %d, want 1", len(frames))
	}
}

func TestStreamParserTwoFramesAndNoise(t *testing.T) {
	first := mustEncode(t, checksum.ModeSum, []byte{0x01})
	second := mustEncode(t, checksum.ModeSum, []byte{0x02})
	parser := NewStreamParser(checksum.ModeSum)

	input := append([]byte{0x00, 0xFF}, first...)
	input = append(input, second...)
	result := parser.Push(input)
	if len(result.Errors) != 0 {
		t.Fatalf("Push() errors = %v", result.Errors)
	}
	if len(result.Frames) != 2 {
		t.Fatalf("frames = %d, want 2", len(result.Frames))
	}
	if !bytes.Equal(result.Frames[0].DataBytes(), []byte{0x01}) || !bytes.Equal(result.Frames[1].DataBytes(), []byte{0x02}) {
		t.Fatalf("unexpected frame data: % X / % X", result.Frames[0].DataBytes(), result.Frames[1].DataBytes())
	}
}

func TestStreamParserInvalidThenValid(t *testing.T) {
	bad := mustEncode(t, checksum.ModeSum, []byte{0x01})
	bad[len(bad)-2] ^= 0xFF
	good := mustEncode(t, checksum.ModeSum, []byte{0x02})
	parser := NewStreamParser(checksum.ModeSum)

	result := parser.Push(append(bad, good...))
	if len(result.Errors) != 1 || !errors.Is(result.Errors[0], ErrInvalidChecksum) {
		t.Fatalf("errors = %v, want one ErrInvalidChecksum", result.Errors)
	}
	if len(result.Frames) != 1 || !bytes.Equal(result.Frames[0].DataBytes(), []byte{0x02}) {
		t.Fatalf("frames = %+v, want valid second frame", result.Frames)
	}
}

func TestStreamParserNoisePartialInvalidThenValid(t *testing.T) {
	bad := mustEncode(t, checksum.ModeSum, []byte{0x01})
	bad[len(bad)-1] = 0x00
	good := mustEncode(t, checksum.ModeSum, []byte{0x03})
	parser := NewStreamParser(checksum.ModeSum)

	first := append([]byte{0x44, 0x55}, bad[:4]...)
	result := parser.Push(first)
	if len(result.Frames) != 0 {
		t.Fatalf("partial invalid push frames = %d, want 0", len(result.Frames))
	}
	result = parser.Push(append(bad[4:], good...))
	if len(result.Errors) == 0 || !errors.Is(result.Errors[0], ErrInvalidEndByte) {
		t.Fatalf("errors = %v, want leading ErrInvalidEndByte", result.Errors)
	}
	if len(result.Frames) != 1 || !bytes.Equal(result.Frames[0].DataBytes(), []byte{0x03}) {
		t.Fatalf("frames = %+v, want valid frame after invalid data", result.Frames)
	}
}

func TestStreamParserMaxFrameSizeExceeded(t *testing.T) {
	parser := NewStreamParser(checksum.ModeSum, WithMaxFrameSize(8))
	result := parser.Push([]byte{StartByte, 0x20, StartByte, 0x00, 0x01})
	if len(result.Errors) == 0 || !errors.Is(result.Errors[0], ErrFrameTooLarge) {
		t.Fatalf("errors = %v, want leading ErrFrameTooLarge", result.Errors)
	}
}

func TestStreamParserCRC16(t *testing.T) {
	raw := mustEncode(t, checksum.ModeCRC16, []byte{0x01})
	parser := NewStreamParser(checksum.ModeCRC16)
	result := parser.Push(raw)
	if len(result.Errors) != 0 || len(result.Frames) != 1 {
		t.Fatalf("Push() result = %+v", result)
	}
}

func mustEncode(t *testing.T, mode checksum.Mode, data []byte) []byte {
	t.Helper()
	raw, err := Encode(New(0x00, 0x01, data), mode)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
