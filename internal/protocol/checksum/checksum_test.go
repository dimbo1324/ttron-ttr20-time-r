package checksum

import "testing"

func TestSum(t *testing.T) {
	got := Sum([]byte{0x00, 0x01, 0x01})
	if got != 0x02 {
		t.Fatalf("Sum() = 0x%02X, want 0x02", got)
	}
}

func TestCRC16Modbus(t *testing.T) {
	got := CRC16Modbus([]byte("123456789"))
	if got != 0x4B37 {
		t.Fatalf("CRC16Modbus() = 0x%04X, want 0x4B37", got)
	}
}
