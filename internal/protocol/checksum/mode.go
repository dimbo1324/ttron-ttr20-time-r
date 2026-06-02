package checksum

import (
	"errors"
	"fmt"
	"strings"
)

type Mode string

const (
	ModeSum   Mode = "sum"
	ModeCRC16 Mode = "crc16"
)

var (
	ErrInvalidMode     = errors.New("invalid checksum mode")
	ErrInvalidChecksum = errors.New("invalid checksum")
)

func ParseMode(raw string) (Mode, error) {
	switch Mode(strings.ToLower(strings.TrimSpace(raw))) {
	case "", ModeSum:
		return ModeSum, nil
	case ModeCRC16:
		return ModeCRC16, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalidMode, raw)
	}
}

func MustParseMode(raw string) Mode {
	mode, err := ParseMode(raw)
	if err != nil {
		panic(err)
	}
	return mode
}

func (m Mode) ChecksumLength() (int, error) {
	switch m {
	case ModeSum:
		return 1, nil
	case ModeCRC16:
		return 2, nil
	default:
		return 0, fmt.Errorf("%w: %q", ErrInvalidMode, m)
	}
}

func Compute(mode Mode, data []byte) ([]byte, error) {
	switch mode {
	case ModeSum:
		return []byte{Sum8(data)}, nil
	case ModeCRC16:
		return CRC16BytesLittleEndian(data), nil
	default:
		return nil, fmt.Errorf("%w: %q", ErrInvalidMode, mode)
	}
}

func Verify(mode Mode, data []byte, expected []byte) error {
	actual, err := Compute(mode, data)
	if err != nil {
		return err
	}
	if len(actual) != len(expected) {
		return fmt.Errorf("%w: mode %s expects %d checksum bytes, got %d", ErrInvalidChecksum, mode, len(actual), len(expected))
	}
	for i := range actual {
		if actual[i] != expected[i] {
			return fmt.Errorf("%w: mode %s", ErrInvalidChecksum, mode)
		}
	}
	return nil
}
