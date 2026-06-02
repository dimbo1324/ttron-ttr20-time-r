package frame

import (
	"encoding/binary"
)

// ComputeSum считает простую сумму байт (mod 256).
func ComputeSum(b []byte) byte {
	var s byte = 0
	for _, x := range b {
		s += x
	}
	return s
}

// ComputeCRC16 считает CRC-16 (Modbus/IBM, poly 0xA001).
func ComputeCRC16(data []byte) uint16 {
	var crc uint16 = 0xFFFF
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&0x0001 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc = crc >> 1
			}
		}
	}
	return crc
}

// VerifyFrame проверяет контрольную сумму фрейма (sum или crc16).
func VerifyFrame(frame []byte) error {
	if len(frame) < 6 {
		return ErrFrameTooShort
	}
	if frame[len(frame)-1] != 0x16 {
		return ErrNoEndByte
	}
	lenByte := int(frame[1])
	payloadStart := 3
	payloadEnd := payloadStart + lenByte
	// Проверяем sum (1 байт)
	if payloadEnd+1 < len(frame) {
		sum := ComputeSum(frame[payloadStart:payloadEnd])
		if frame[payloadEnd] == sum {
			return nil
		}
	}
	// Проверяем crc16 (2 байта little-endian)
	if payloadEnd+2 < len(frame) {
		got := binary.LittleEndian.Uint16(frame[payloadEnd : payloadEnd+2])
		crc := ComputeCRC16(frame[payloadStart:payloadEnd])
		if got == crc {
			return nil
		}
	}
	return ErrChecksumMismatch
}

// CorruptChecksum испортит байты контрольной суммы (для тестов).
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

var (
	ErrFrameTooShort    = &FrameError{"frame too short"}
	ErrNoEndByte        = &FrameError{"no end byte 0x16"}
	ErrChecksumMismatch = &FrameError{"checksum mismatch"}
)

type FrameError struct{ s string }

func (e *FrameError) Error() string { return e.s }
