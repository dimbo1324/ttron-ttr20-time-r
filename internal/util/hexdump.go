package util

import "fmt"

func HexDump(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	out := make([]byte, 0, len(data)*3)
	for i, b := range data {
		if i > 0 {
			out = append(out, ' ')
		}
		out = fmt.Appendf(out, "%02X", b)
	}
	return string(out)
}
