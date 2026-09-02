package frame

import (
	"errors"
	"testing"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
)

func TestStreamParserSkipsLeadingNoise(t *testing.T) {
	raw := buildFrame(t, checksum.ModeSum, 0x00, 0x01, []byte{0x01})
	parser := NewStreamParser(checksum.ModeSum)

	result := parser.Push(append([]byte{0xAA, 0xBB, 0xCC}, raw...))
	if len(result.Frames) != 1 {
		t.Fatalf("Push() returned %d frames, want 1", len(result.Frames))
	}
	if parser.BufferedLen() != 0 {
		t.Fatalf("BufferedLen() = %d, want 0", parser.BufferedLen())
	}
}

func TestStreamParserResyncsAfterInvalidRepeatedStartByte(t *testing.T) {
	raw := buildFrame(t, checksum.ModeSum, 0x00, 0x01, []byte{0x01})
	broken := []byte{StartByte, 0x03, 0x00, 0x00, 0x01, 0x01, 0x02, EndByte}

	parser := NewStreamParser(checksum.ModeSum)
	result := parser.Push(append(broken, raw...))

	if len(result.Errors) == 0 {
		t.Fatal("Push() must report the malformed frame")
	}
	if !errors.Is(result.Errors[0], ErrInvalidRepeatedStartByte) {
		t.Fatalf("Push() error = %v, want %v", result.Errors[0], ErrInvalidRepeatedStartByte)
	}
	if len(result.Frames) != 1 {
		t.Fatalf("Push() returned %d frames, want the parser to resynchronize", len(result.Frames))
	}
}

func TestStreamParserResyncsAfterInvalidLength(t *testing.T) {
	raw := buildFrame(t, checksum.ModeSum, 0x00, 0x01, []byte{0x01})
	broken := []byte{StartByte, 0x01, StartByte, 0x00, 0x01, EndByte}

	parser := NewStreamParser(checksum.ModeSum)
	result := parser.Push(append(broken, raw...))

	if len(result.Errors) == 0 || !errors.Is(result.Errors[0], ErrInvalidLength) {
		t.Fatalf("Push() errors = %v, want %v", result.Errors, ErrInvalidLength)
	}
	if len(result.Frames) != 1 {
		t.Fatalf("Push() returned %d frames, want the parser to recover", len(result.Frames))
	}
}

func TestStreamParserResyncsAfterBadChecksum(t *testing.T) {
	good := buildFrame(t, checksum.ModeSum, 0x00, 0x01, []byte{0x01})
	corrupt := append([]byte(nil), good...)
	corrupt[len(corrupt)-2] ^= 0xFF

	parser := NewStreamParser(checksum.ModeSum)
	result := parser.Push(append(corrupt, good...))

	if len(result.Errors) == 0 || !errors.Is(result.Errors[0], ErrInvalidChecksum) {
		t.Fatalf("Push() errors = %v, want %v", result.Errors, ErrInvalidChecksum)
	}
	if len(result.Frames) != 1 {
		t.Fatalf("Push() returned %d frames, want the following frame", len(result.Frames))
	}
}

func TestStreamParserDropsUnrecoverableBuffer(t *testing.T) {
	parser := NewStreamParser(checksum.ModeSum)

	result := parser.Push([]byte{StartByte, 0x03, 0x00, 0x00, 0x01, 0x01, 0x02, 0x00})
	if len(result.Errors) == 0 {
		t.Fatal("Push() must report the malformed frame")
	}
	if len(result.Frames) != 0 {
		t.Fatalf("Push() returned %d frames, want 0", len(result.Frames))
	}
	if parser.BufferedLen() != 0 {
		t.Fatalf("BufferedLen() = %d, want the buffer dropped when no start byte remains", parser.BufferedLen())
	}
}

func TestStreamParserRejectsOversizedFrame(t *testing.T) {
	parser := NewStreamParser(checksum.ModeSum, WithMaxFrameSize(8))

	result := parser.Push([]byte{StartByte, 0xFF, StartByte, 0x00, 0x01})
	if len(result.Errors) == 0 {
		t.Fatal("Push() must reject a frame above the size limit")
	}
	if !errors.Is(result.Errors[0], ErrFrameTooLarge) {
		t.Fatalf("Push() error = %v, want %v", result.Errors[0], ErrFrameTooLarge)
	}
}

func TestStreamParserGuardsAgainstUnboundedGrowth(t *testing.T) {
	parser := NewStreamParser(checksum.ModeSum, WithMaxFrameSize(16))

	noise := make([]byte, 64)
	for i := range noise {
		noise[i] = StartByte
	}
	result := parser.Push(noise)

	if parser.BufferedLen() > 64 {
		t.Fatalf("BufferedLen() = %d, want a bounded buffer", parser.BufferedLen())
	}
	if len(result.Frames) != 0 {
		t.Fatalf("Push() returned %d frames, want 0", len(result.Frames))
	}
}

func TestStreamParserHandlesEmptyPush(t *testing.T) {
	parser := NewStreamParser(checksum.ModeSum)

	result := parser.Push(nil)
	if len(result.Frames) != 0 || len(result.Errors) != 0 {
		t.Fatalf("Push(nil) = %+v, want an empty result", result)
	}
}

func TestStreamParserResyncToNextStart(t *testing.T) {
	parser := NewStreamParser(checksum.ModeSum)
	parser.buffer = []byte{StartByte, 0xAA, 0xBB, StartByte, 0xCC}

	parser.resyncToNextStart()
	if parser.BufferedLen() != 2 {
		t.Fatalf("BufferedLen() = %d, want 2", parser.BufferedLen())
	}
	if parser.buffer[0] != StartByte {
		t.Fatalf("buffer = % X, want it to start at the next start byte", parser.buffer)
	}

	parser.buffer = []byte{StartByte, 0xAA, 0xBB}
	parser.resyncToNextStart()
	if parser.BufferedLen() != 0 {
		t.Fatalf("BufferedLen() = %d, want the buffer dropped", parser.BufferedLen())
	}
}

func TestIndexByte(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		target byte
		want   int
	}{
		{name: "found", data: []byte{0x01, 0x02, 0x03}, target: 0x02, want: 1},
		{name: "first byte", data: []byte{0x01, 0x02}, target: 0x01, want: 0},
		{name: "absent", data: []byte{0x01, 0x02}, target: 0x09, want: -1},
		{name: "empty", data: nil, target: 0x01, want: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := indexByte(tt.data, tt.target); got != tt.want {
				t.Fatalf("indexByte() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestWithMaxFrameSizeIgnoresNonPositive(t *testing.T) {
	parser := NewStreamParser(checksum.ModeSum, WithMaxFrameSize(0), WithMaxFrameSize(-1))
	if parser.maxFrameSize != DefaultMaxFrameSize {
		t.Fatalf("maxFrameSize = %d, want %d", parser.maxFrameSize, DefaultMaxFrameSize)
	}
}
