package command

import (
	"errors"
	"fmt"
)

type ID byte

const (
	ReadTime     ID = 0x01
	ReadIdentity ID = 0x02
)

const (
	NameReadTime     = "read-time"
	NameReadIdentity = "read-identity"
)

var (
	ErrEmptyPayload      = errors.New("empty command payload")
	ErrUnexpectedCommand = errors.New("unexpected command")
	ErrInvalidPayload    = errors.New("invalid command payload")
	ErrInvalidTime       = errors.New("invalid read-time timestamp")
	ErrInvalidIdentity   = errors.New("invalid identity payload")
	ErrUnknownCommand    = errors.New("unknown command")
	ErrDuplicateCommand  = errors.New("duplicate command registration")
	ErrInvalidDescriptor = errors.New("invalid command descriptor")
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
