package emulator

import (
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
)

func BuildTimeResponse(reqCtrl, reqAddr byte, _ []byte, crcMode string, _ byte) []byte {
	payload := append([]byte{0x01}, []byte(time.Now().Format("2006-01-02 15:04:05"))...)
	return frame.AppendChecksum(frame.BuildSkeleton(reqCtrl|0x80, reqAddr, payload), crcMode)
}

func BuildAckResponse(reqCtrl, reqAddr byte, reqData []byte, crcMode string, _ byte) []byte {
	cmd := byte(0xFF)
	if len(reqData) > 0 {
		cmd = reqData[0]
	}
	payload := append([]byte{cmd}, []byte("OK")...)
	return frame.AppendChecksum(frame.BuildSkeleton(reqCtrl|0x80, reqAddr, payload), crcMode)
}
