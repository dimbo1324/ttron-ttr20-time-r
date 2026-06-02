package codec

import (
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/command"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
)

type Codec struct {
	Mode    checksum.Mode
	Control byte
	Address byte
}

func New(mode checksum.Mode, control, address byte) Codec {
	return Codec{Mode: mode, Control: control, Address: address}
}

func (c Codec) EncodeReadTimeRequest() ([]byte, error) {
	return frame.Encode(frame.New(c.Control, c.Address, command.BuildReadTimeRequest()), c.Mode)
}

func (c Codec) DecodeReadTimeRequest(raw []byte) (frame.Frame, error) {
	f, err := frame.Decode(raw, c.Mode)
	if err != nil {
		return frame.Frame{}, err
	}
	if err := command.ParseReadTimeRequest(f.DataBytes()); err != nil {
		return frame.Frame{}, err
	}
	return f, nil
}

func (c Codec) EncodeReadTimeResponse(request frame.Frame, t time.Time) ([]byte, error) {
	return frame.Encode(frame.New(request.Control|0x80, request.Address, command.BuildReadTimeResponse(t)), c.Mode)
}

func (c Codec) DecodeReadTimeResponse(raw []byte) (frame.Frame, command.ReadTimeResponse, error) {
	f, err := frame.Decode(raw, c.Mode)
	if err != nil {
		return frame.Frame{}, command.ReadTimeResponse{}, err
	}
	resp, err := command.ParseReadTimeResponse(f.DataBytes())
	if err != nil {
		return frame.Frame{}, command.ReadTimeResponse{}, err
	}
	return f, resp, nil
}

func (c Codec) EncodeACK(request frame.Frame, data []byte) ([]byte, error) {
	cmd := byte(0xFF)
	if len(data) > 0 {
		cmd = data[0]
	}
	payload := append([]byte{cmd}, []byte("OK")...)
	return frame.Encode(frame.New(request.Control|0x80, request.Address, payload), c.Mode)
}
