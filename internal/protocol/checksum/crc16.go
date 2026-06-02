package checksum

import "encoding/binary"

func CRC16(data []byte) uint16 {
	var crc uint16 = 0xFFFF
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&0x0001 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

func CRC16BytesLittleEndian(data []byte) []byte {
	out := make([]byte, 2)
	binary.LittleEndian.PutUint16(out, CRC16(data))
	return out
}

// CRC16Modbus is kept as a compatibility alias for the Step 1 baseline API.
func CRC16Modbus(data []byte) uint16 {
	return CRC16(data)
}
