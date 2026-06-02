package frame

import (
	"errors"
	"fmt"
)

var (
	ErrFrameTooShort              = errors.New("frame too short")
	ErrInvalidStartByte           = errors.New("invalid start byte")
	ErrInvalidRepeatedStartByte   = errors.New("invalid repeated start byte")
	ErrInvalidLength              = errors.New("invalid length")
	ErrInvalidChecksum            = errors.New("invalid checksum")
	ErrInvalidEndByte             = errors.New("invalid end byte")
	ErrFrameTooLarge              = errors.New("frame too large")
	ErrUnsupportedPayloadLength   = errors.New("unsupported payload length")
	ErrInvalidControlAddressBytes = errors.New("frame must contain control and address bytes")
)

func wrapError(base error, format string, args ...any) error {
	return fmt.Errorf("%w: %s", base, fmt.Sprintf(format, args...))
}
