package frame

import (
	"bytes"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
)

// CRC16 is kept for callers that still compare mode strings from Step 1.
const CRC16 = string(checksum.ModeCRC16)

func BuildSkeleton(control, addr byte, data []byte) []byte {
	f := New(control, addr, data)
	out := []byte{StartByte, byte(len(f.PayloadBytes())), StartByte}
	out = append(out, f.PayloadBytes()...)
	return out
}

func AppendChecksum(raw []byte, rawMode string) []byte {
	mode, err := checksum.ParseMode(rawMode)
	if err != nil || len(raw) < 3 {
		return append([]byte(nil), raw...)
	}
	sum, err := checksum.Compute(mode, raw[3:])
	if err != nil {
		return append([]byte(nil), raw...)
	}
	out := append([]byte(nil), raw...)
	out = append(out, sum...)
	out = append(out, EndByte)
	return out
}

func Verify(raw []byte) error {
	if _, err := Decode(raw, checksum.ModeSum); err == nil {
		return nil
	}
	if _, err := Decode(raw, checksum.ModeCRC16); err == nil {
		return nil
	}
	return ErrInvalidChecksum
}

func ExtractFrame(buf *bytes.Buffer) ([]byte, bool) {
	parser := NewStreamParser(checksum.ModeSum)
	result := parser.Push(buf.Bytes())
	if len(result.Frames) > 0 {
		raw := result.Frames[0].RawBytes()
		buf.Next(len(raw))
		return raw, true
	}

	parser = NewStreamParser(checksum.ModeCRC16)
	result = parser.Push(buf.Bytes())
	if len(result.Frames) > 0 {
		raw := result.Frames[0].RawBytes()
		buf.Next(len(raw))
		return raw, true
	}
	return nil, false
}

func CorruptChecksum(raw []byte, rawMode string) {
	mode, err := checksum.ParseMode(rawMode)
	if err != nil {
		mode = checksum.ModeSum
	}
	checksumLen, err := mode.ChecksumLength()
	if err != nil || len(raw) < checksumLen+1 {
		return
	}
	raw[len(raw)-1-checksumLen] ^= 0xFF
}
