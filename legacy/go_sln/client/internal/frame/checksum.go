package frame

// ComputeSum считает простую сумму байт (mod 256).
func ComputeSum(b []byte) byte {
	var s byte = 0
	for _, v := range b {
		s += v
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
