package command

import (
	"errors"
	"fmt"
)

type ID byte

const (
	ReadTime ID = 0x01
)

var (
	ErrEmptyPayload      = errors.New("empty command payload")
	ErrUnexpectedCommand = errors.New("unexpected command")
	ErrInvalidPayload    = errors.New("invalid command payload")
	ErrInvalidTime       = errors.New("invalid read-time timestamp")
)

func ParseID(data []byte) (ID, error) {
	if len(data) == 0 {
		return 0, ErrEmptyPayload
	}
	return ID(data[0]), nil
}

func Expect(data []byte, id ID) error {
	got, err := ParseID(data)
	if err != nil {
		return err
	}
	if got != id {
		return fmt.Errorf("%w: got 0x%02X, want 0x%02X", ErrUnexpectedCommand, byte(got), byte(id))
	}
	return nil
}
