package frame

import (
	"bytes"
	"encoding/binary"
)

// ExtractFrame ищет и извлекает первый полный фрейм из буфера
// Формат фрейма: 0x68 | LEN | 0x68 | CONTROL | ADDR | DATA... | CHECKSUM | 0x16
func ExtractFrame(buf *bytes.Buffer) ([]byte, bool) {
	b := buf.Bytes()
	start := bytes.IndexByte(b, 0x68)
	if start < 0 {
		return nil, false
	}
	if len(b) <= start+2 {
		return nil, false
	}
	if b[start+2] != 0x68 {
		// Сдвигаем буфер, чтобы не застрять на неверном байте
		buf.Next(start + 1)
		return nil, false
	}
	lenByte := int(b[start+1])
	// минимальная полная длина с 1-байтной контрольной суммой
	minEnd := start + 3 + lenByte + 1 + 1
	if len(b) < minEnd {
		return nil, false
	}
	endIdx1 := start + 3 + lenByte + 1
	if endIdx1+1 < len(b) && b[endIdx1+1] == 0x16 {
		frame := make([]byte, endIdx1-start+2)
		copy(frame, b[start:endIdx1+2])
		buf.Next(endIdx1 + 2)
		return frame, true
	}
	// пробуем вариант с 2-байтной CRC
	endIdx2 := start + 3 + lenByte + 2
	if endIdx2+1 < len(b) && b[endIdx2+1] == 0x16 {
		frame := make([]byte, endIdx2-start+2)
		copy(frame, b[start:endIdx2+2])
		buf.Next(endIdx2 + 2)
		return frame, true
	}
	return nil, false
}

// PayloadData возвращает DATA (без control и addr) из фрейма
func PayloadData(frame []byte) []byte {
	if len(frame) <= 5 {
		return nil
	}
	lenByte := int(frame[1])
	if lenByte < 2 {
		return nil
	}
	dataLen := lenByte - 2
	dataStart := 5
	if dataStart+dataLen > len(frame)-3 {
		if dataStart >= len(frame)-3 {
			return nil
		}
		return frame[dataStart : len(frame)-3]
	}
	return frame[dataStart : dataStart+dataLen]
}

// BuildSkeleton формирует начальную часть фрейма (без CRC и 0x16).
func BuildSkeleton(control byte, addr byte, data []byte) []byte {
	lenByte := byte(2 + len(data))
	var b bytes.Buffer
	b.WriteByte(0x68)
	b.WriteByte(lenByte)
	b.WriteByte(0x68)
	b.WriteByte(control)
	b.WriteByte(addr)
	b.Write(data)
	return b.Bytes()
}

// AppendChecksum дописывает checksum (sum или crc16) и 0x16
func AppendChecksum(frameSoFar []byte, crcMode string) []byte {
	if crcMode == "crc16" {
		crc := ComputeCRC16(frameSoFar[3:])
		tmp := make([]byte, 2)
		binary.LittleEndian.PutUint16(tmp, crc)
		return append(frameSoFar, append(tmp, 0x16)...)
	}
	sum := ComputeSum(frameSoFar[3:])
	return append(frameSoFar, append([]byte{sum}, 0x16)...)
}

// VerifyFrame проверяет конец 0x16 и checksum (sum или crc16).
func VerifyFrame(frame []byte) error {
	if len(frame) < 6 {
		return ErrFrameTooShort
	}
	if frame[len(frame)-1] != 0x16 {
		return ErrNoEndByte
	}
	payloadStart := 3
	payloadEnd := payloadStart + int(frame[1])
	if payloadEnd > len(frame)-3 {
		payloadEnd = len(frame) - 3
	}
	// try sum
	if payloadEnd+1 < len(frame) {
		if frame[payloadEnd] == ComputeSum(frame[payloadStart:payloadEnd]) {
			return nil
		}
	}
	// try crc16
	if payloadEnd+2 < len(frame) {
		got := binary.LittleEndian.Uint16(frame[payloadEnd : payloadEnd+2])
		if got == ComputeCRC16(frame[payloadStart:payloadEnd]) {
			return nil
		}
	}
	return ErrChecksumMismatch
}

// CorruptChecksum ломает checksum для теста
func CorruptChecksum(frame []byte, crcMode string) {
	if crcMode == "crc16" {
		if len(frame) >= 4 {
			idx := len(frame) - 3
			frame[idx] ^= 0x01
		}
	} else {
		if len(frame) >= 3 {
			idx := len(frame) - 2
			frame[idx] ^= 0xFF
		}
	}
}

type FrameError struct{ s string }

func (e *FrameError) Error() string { return e.s }

var (
	ErrFrameTooShort    = &FrameError{"frame too short"}
	ErrNoEndByte        = &FrameError{"no end byte 0x16"}
	ErrChecksumMismatch = &FrameError{"checksum mismatch"}
)
