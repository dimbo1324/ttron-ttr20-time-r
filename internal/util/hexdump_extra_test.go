package util

import "testing"

func TestHexDumpEmptyInput(t *testing.T) {
	if got := HexDump(nil); got != "" {
		t.Fatalf("HexDump(nil) = %q, want an empty string", got)
	}
	if got := HexDump([]byte{}); got != "" {
		t.Fatalf("HexDump([]) = %q, want an empty string", got)
	}
}

func TestHexDumpSingleByte(t *testing.T) {
	if got := HexDump([]byte{0x0F}); got != "0F" {
		t.Fatalf("HexDump() = %q, want %q", got, "0F")
	}
}
