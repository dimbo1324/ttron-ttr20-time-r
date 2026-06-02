package frame

import (
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
)

const DefaultMaxFrameSize = 1024

type StreamParser struct {
	mode         checksum.Mode
	maxFrameSize int
	buffer       []byte
}

type Option func(*StreamParser)

func WithMaxFrameSize(size int) Option {
	return func(p *StreamParser) {
		if size > 0 {
			p.maxFrameSize = size
		}
	}
}

type ParseResult struct {
	Frames []Frame
	Errors []error
}

func NewStreamParser(mode checksum.Mode, opts ...Option) *StreamParser {
	p := &StreamParser{
		mode:         mode,
		maxFrameSize: DefaultMaxFrameSize,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *StreamParser) Push(data []byte) ParseResult {
	var result ParseResult
	if len(data) > 0 {
		p.buffer = append(p.buffer, data...)
	}

	checksumLen, err := p.mode.ChecksumLength()
	if err != nil {
		result.Errors = append(result.Errors, err)
		p.buffer = nil
		return result
	}

	for {
		if len(p.buffer) == 0 {
			return result
		}
		if len(p.buffer) > p.maxFrameSize {
			result.Errors = append(result.Errors, wrapError(ErrFrameTooLarge, "buffer has %d bytes", len(p.buffer)))
			p.resyncToNextStart()
			if len(p.buffer) > p.maxFrameSize {
				p.buffer = nil
				return result
			}
		}

		start := indexByte(p.buffer, StartByte)
		if start < 0 {
			p.buffer = nil
			return result
		}
		if start > 0 {
			p.buffer = p.buffer[start:]
		}

		if len(p.buffer) < 3 {
			return result
		}
		if p.buffer[2] != StartByte {
			result.Errors = append(result.Errors, wrapError(ErrInvalidRepeatedStartByte, "got 0x%02X", p.buffer[2]))
			p.buffer = p.buffer[1:]
			continue
		}

		payloadLen := int(p.buffer[1])
		if payloadLen < 2 {
			result.Errors = append(result.Errors, wrapError(ErrInvalidLength, "payload length %d", payloadLen))
			p.buffer = p.buffer[1:]
			continue
		}

		wantLen := 3 + payloadLen + checksumLen + 1
		if wantLen > p.maxFrameSize {
			result.Errors = append(result.Errors, wrapError(ErrFrameTooLarge, "frame wants %d bytes", wantLen))
			p.buffer = p.buffer[1:]
			continue
		}
		if len(p.buffer) < wantLen {
			return result
		}

		candidate := append([]byte(nil), p.buffer[:wantLen]...)
		decoded, err := Decode(candidate, p.mode)
		if err != nil {
			result.Errors = append(result.Errors, err)
			if candidate[len(candidate)-1] == EndByte {
				p.buffer = p.buffer[wantLen:]
			} else {
				p.buffer = p.buffer[1:]
			}
			continue
		}

		result.Frames = append(result.Frames, decoded)
		p.buffer = p.buffer[wantLen:]
	}
}

func (p *StreamParser) BufferedLen() int {
	return len(p.buffer)
}

func (p *StreamParser) resyncToNextStart() {
	start := indexByte(p.buffer[1:], StartByte)
	if start < 0 {
		p.buffer = nil
		return
	}
	p.buffer = p.buffer[start+1:]
}

func indexByte(data []byte, target byte) int {
	for i, b := range data {
		if b == target {
			return i
		}
	}
	return -1
}
