package util

import "testing"

func TestHexDump(t *testing.T) {
	got := HexDump([]byte{0x68, 0x03, 0x16})
	if got != "68 03 16" {
		t.Fatalf("HexDump() = %q, want %q", got, "68 03 16")
	}
}
