package schedule

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type Mode string

const (
	ModeInterval Mode = "interval"
	ModeAligned  Mode = "aligned"
)

var (
	ErrInvalidMode     = errors.New("invalid schedule mode")
	ErrInvalidInterval = errors.New("schedule interval must be positive")
	ErrInvalidOffset   = errors.New("schedule offset must be non-negative and below the interval")
)

type Schedule interface {
	Next(now time.Time) time.Time
	Mode() Mode
	Interval() time.Duration
	String() string
}

func ParseMode(value string) (Mode, error) {
	switch Mode(strings.ToLower(strings.TrimSpace(value))) {
	case ModeInterval, "":
		return ModeInterval, nil
	case ModeAligned:
		return ModeAligned, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalidMode, value)
	}
}

func New(mode Mode, interval, offset time.Duration) (Schedule, error) {
	if interval <= 0 {
		return nil, ErrInvalidInterval
	}
	switch mode {
	case ModeInterval, "":
		return NewInterval(interval)
	case ModeAligned:
		return NewAligned(interval, offset)
	default:
		return nil, fmt.Errorf("%w: %q", ErrInvalidMode, mode)
	}
}
