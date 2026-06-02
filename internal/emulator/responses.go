package emulator

import (
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/codec"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
)

func BuildTimeResponse(reqCtrl, reqAddr byte, _ []byte, crcMode string, _ byte) []byte {
	mode, err := checksum.ParseMode(crcMode)
	if err != nil {
		mode = checksum.ModeSum
	}
	req := frame.New(reqCtrl, reqAddr, nil)
	raw, err := codec.New(mode, reqCtrl, reqAddr).EncodeReadTimeResponse(req, time.Now())
	if err != nil {
		return nil
	}
	return raw
}

func BuildAckResponse(reqCtrl, reqAddr byte, reqData []byte, crcMode string, _ byte) []byte {
	mode, err := checksum.ParseMode(crcMode)
	if err != nil {
		mode = checksum.ModeSum
	}
	req := frame.New(reqCtrl, reqAddr, reqData)
	raw, err := codec.New(mode, reqCtrl, reqAddr).EncodeACK(req, reqData)
	if err != nil {
		return nil
	}
	return raw
}
