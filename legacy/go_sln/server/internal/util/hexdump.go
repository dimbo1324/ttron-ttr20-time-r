package util

import "fmt"

func HexDump(b []byte) string {
	if b == nil || len(b) == 0 {
		return ""
	}
	s := make([]byte, 0, len(b)*3)
	for i, v := range b {
		if i > 0 {
			s = append(s, ' ')
		}
		s = append(s, hexByte(v)...)
	}
	return string(s)
}

func hexByte(b byte) []byte {
	return []byte(fmt.Sprintf("%02X", b))
}
