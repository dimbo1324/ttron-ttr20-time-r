package frame

import (
	"bytes"
	"encoding/binary"
)

// ExtractFrame ищет и извлекает первый полный фрейм из буфера
// Возвращает копию фрейма и true, если найденный фрейм удалён из буфера
func ExtractFrame(buf *bytes.Buffer) ([]byte, bool) {
	b := buf.Bytes()

	start := bytes.IndexByte(b, 0x68)
	if start < 0 {
		return nil, false
	}

	// Нужны минимум: 0x68, LEN, 0x68
	if len(b) < start+3 {
		return nil, false
	}

	// Проверяем второй 0x68
	if b[start+2] != 0x68 {
		buf.Next(start + 1)
		return nil, false
	}

	lenByte := int(b[start+1])
	if lenByte < 0 {
		buf.Next(start + 1)
		return nil, false
	}

	payloadStart := start + 3
	payloadEnd := payloadStart + lenByte

	// Убедимся, что в буфере есть хотя бы payload + 1 байт checksum + 0x16
	if len(b) < payloadEnd+2 {
		return nil, false
	}

	// Проверка на 1-байтный checksum и окончание 0x16
	endIdx1 := payloadEnd + 1
	if endIdx1 < len(b) && b[endIdx1] == 0x16 {
		frame := make([]byte, endIdx1-start+1)
		copy(frame, b[start:endIdx1+1])
		buf.Next(endIdx1 + 1)
		return frame, true
	}

	// Проверка на 2-байтный checksum и окончание 0x16
	endIdx2 := payloadEnd + 2
	if endIdx2 < len(b) && b[endIdx2] == 0x16 {
		frame := make([]byte, endIdx2-start+1)
		copy(frame, b[start:endIdx2+1])
		buf.Next(endIdx2 + 1)
		return frame, true
	}

	return nil, false
}

// PayloadData возвращает DATA (без CONTROL и ADDR) из фрейма
// Использует LEN (frame[1) для корректного вычисления границ
func PayloadData(frame []byte) []byte {
	if len(frame) < 7 {
		return nil
	}
	lenByte := int(frame[1])
	if lenByte < 2 {
		return nil
	}
	payloadStart := 3
	payloadEnd := payloadStart + lenByte

	// Должно оставаться минимум checksum + 0x16
	if payloadEnd > len(frame)-2 {
		return nil
	}

	// DATA начинается после CONTROL и ADDR
	dataStart := payloadStart + 2
	if dataStart > payloadEnd {
		return nil
	}
	return frame[dataStart:payloadEnd]
}

// BuildSkeleton формирует базовую часть кадра без CRC и 0x16
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

// AppendChecksum добавляет CRC/SUM и терминатор 0x16
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
