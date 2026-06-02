package util

import "fmt"

// HexDump возвращает компактный hex-дамп (строкой) для логов
func HexDump(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	out := ""
	for i, v := range b {
		if i > 0 {
			out += " "
		}
		out += fmt.Sprintf("%02X", v)
	}
	return out
}
