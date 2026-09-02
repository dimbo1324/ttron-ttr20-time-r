package schedule

import (
	"fmt"
	"time"
)

type AlignedSchedule struct {
	interval time.Duration
	offset   time.Duration
}

func NewAligned(interval, offset time.Duration) (AlignedSchedule, error) {
	if interval <= 0 {
		return AlignedSchedule{}, ErrInvalidInterval
	}
	if offset < 0 || offset >= interval {
		return AlignedSchedule{}, ErrInvalidOffset
	}
	return AlignedSchedule{interval: interval, offset: offset}, nil
}

func (s AlignedSchedule) Next(now time.Time) time.Time {
	interval := int64(s.interval)
	shifted := now.Add(-s.offset).UnixNano()

	boundary := shifted - floorMod(shifted, interval) + interval
	next := time.Unix(0, boundary).Add(s.offset).In(now.Location())
	if !next.After(now) {
		next = next.Add(s.interval)
	}
	return next
}

func (s AlignedSchedule) Mode() Mode {
	return ModeAligned
}

func (s AlignedSchedule) Interval() time.Duration {
	return s.interval
}

func (s AlignedSchedule) Offset() time.Duration {
	return s.offset
}

func (s AlignedSchedule) String() string {
	if s.offset == 0 {
		return fmt.Sprintf("aligned every %s", s.interval)
	}
	return fmt.Sprintf("aligned every %s at offset %s", s.interval, s.offset)
}

func floorMod(value, modulus int64) int64 {
	remainder := value % modulus
	if remainder < 0 {
		remainder += modulus
	}
	return remainder
}
