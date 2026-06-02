package frame

import (
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
)

func Encode(f Frame, mode checksum.Mode) ([]byte, error) {
	payload := f.PayloadBytes()
	if len(payload) < 2 {
		return nil, ErrInvalidControlAddressBytes
	}
	if len(payload) > 255 {
		return nil, wrapError(ErrUnsupportedPayloadLength, "%d bytes", len(payload))
	}
	sum, err := checksum.Compute(mode, payload)
	if err != nil {
		return nil, err
	}

	out := make([]byte, 0, 3+len(payload)+len(sum)+1)
	out = append(out, StartByte, byte(len(payload)), StartByte)
	out = append(out, payload...)
	out = append(out, sum...)
	out = append(out, EndByte)
	return out, nil
}
