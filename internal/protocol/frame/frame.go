package frame

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
)

const (
	StartByte = 0x68
	EndByte   = 0x16
	CRC16     = "crc16"
)

var (
	ErrFrameTooShort    = errors.New("frame too short")
	ErrNoEndByte        = errors.New("no end byte 0x16")
	ErrChecksumMismatch = errors.New("checksum mismatch")
)

// ExtractFrame extracts the first complete FT1.2-like frame from buf.
//
// Current active format:
// 0x68 | LEN | 0x68 | CONTROL | ADDR | DATA... | CHECKSUM | 0x16
func ExtractFrame(buf *bytes.Buffer) ([]byte, bool) {
	b := buf.Bytes()
	start := bytes.IndexByte(b, StartByte)
	if start < 0 {
		buf.Reset()
		return nil, false
	}
	if len(b) < start+3 {
		return nil, false
	}
	if b[start+2] != StartByte {
		buf.Next(start + 1)
		return nil, false
	}

	payloadStart := start + 3
	payloadEnd := payloadStart + int(b[start+1])
	if len(b) < payloadEnd+2 {
		return nil, false
	}

	endIdx1 := payloadEnd + 1
	if endIdx1 < len(b) && b[endIdx1] == EndByte {
		out := make([]byte, endIdx1-start+1)
		copy(out, b[start:endIdx1+1])
		buf.Next(endIdx1 + 1)
		return out, true
	}

	endIdx2 := payloadEnd + 2
	if endIdx2 < len(b) && b[endIdx2] == EndByte {
		out := make([]byte, endIdx2-start+1)
		copy(out, b[start:endIdx2+1])
		buf.Next(endIdx2 + 1)
		return out, true
	}

	return nil, false
}

func PayloadData(f []byte) []byte {
	if len(f) < 7 {
		return nil
	}
	payloadEnd := 3 + int(f[1])
	dataStart := 5
	if int(f[1]) < 2 || payloadEnd > len(f)-2 || dataStart > payloadEnd {
		return nil
	}
	return f[dataStart:payloadEnd]
}

func BuildSkeleton(control, addr byte, data []byte) []byte {
	var b bytes.Buffer
	b.WriteByte(StartByte)
	b.WriteByte(byte(2 + len(data)))
	b.WriteByte(StartByte)
	b.WriteByte(control)
	b.WriteByte(addr)
	b.Write(data)
	return b.Bytes()
}

func AppendChecksum(f []byte, mode string) []byte {
	if mode == CRC16 {
		crc := checksum.CRC16Modbus(f[3:])
		tmp := make([]byte, 2)
		binary.LittleEndian.PutUint16(tmp, crc)
		return append(f, tmp[0], tmp[1], EndByte)
	}
	return append(f, checksum.Sum(f[3:]), EndByte)
}

func Verify(f []byte) error {
	if len(f) < 6 {
		return ErrFrameTooShort
	}
	if f[len(f)-1] != EndByte {
		return ErrNoEndByte
	}

	payloadStart := 3
	payloadEnd := payloadStart + int(f[1])
	if payloadEnd > len(f)-2 {
		return ErrFrameTooShort
	}

	if payloadEnd+1 < len(f) && f[payloadEnd] == checksum.Sum(f[payloadStart:payloadEnd]) {
		return nil
	}
	if payloadEnd+2 < len(f) {
		got := binary.LittleEndian.Uint16(f[payloadEnd : payloadEnd+2])
		if got == checksum.CRC16Modbus(f[payloadStart:payloadEnd]) {
			return nil
		}
	}
	return ErrChecksumMismatch
}

func CorruptChecksum(f []byte, mode string) {
	if mode == CRC16 {
		if len(f) >= 4 {
			f[len(f)-3] ^= 0x01
		}
		return
	}
	if len(f) >= 3 {
		f[len(f)-2] ^= 0xFF
	}
}
