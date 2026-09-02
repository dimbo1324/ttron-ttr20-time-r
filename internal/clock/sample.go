package clock

import (
	"errors"
	"time"
)

var (
	ErrInvalidSample = errors.New("invalid clock sample")
	ErrEmptyWindow   = errors.New("clock window is empty")
)

type Sample struct {
	RequestedAt time.Time
	ReceivedAt  time.Time
	DeviceTime  time.Time
}

func (s Sample) Validate() error {
	if s.RequestedAt.IsZero() || s.ReceivedAt.IsZero() || s.DeviceTime.IsZero() {
		return ErrInvalidSample
	}
	if s.ReceivedAt.Before(s.RequestedAt) {
		return ErrInvalidSample
	}
	return nil
}

func (s Sample) RoundTrip() time.Duration {
	if s.RequestedAt.IsZero() || s.ReceivedAt.IsZero() {
		return 0
	}
	rtt := s.ReceivedAt.Sub(s.RequestedAt)
	if rtt < 0 {
		return 0
	}
	return rtt
}

func (s Sample) Reference() time.Time {
	return s.RequestedAt.Add(s.RoundTrip() / 2)
}

func (s Sample) Skew() time.Duration {
	return s.DeviceTime.Sub(s.Reference())
}

func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}
