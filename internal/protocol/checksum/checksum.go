package checksum

// Sum computes an 8-bit additive checksum over the provided bytes.
func Sum(data []byte) byte {
	var sum byte
	for _, b := range data {
		sum += b
	}
	return sum
}

// CRC16Modbus computes CRC-16/Modbus with polynomial 0xA001.
func CRC16Modbus(data []byte) uint16 {
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
