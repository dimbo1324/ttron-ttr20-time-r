package schedule

import (
	"fmt"
	"time"
)

type IntervalSchedule struct {
	interval time.Duration
}

func NewInterval(interval time.Duration) (IntervalSchedule, error) {
	if interval <= 0 {
		return IntervalSchedule{}, ErrInvalidInterval
	}
	return IntervalSchedule{interval: interval}, nil
}

func (s IntervalSchedule) Next(now time.Time) time.Time {
	return now.Add(s.interval)
}

func (s IntervalSchedule) Mode() Mode {
	return ModeInterval
}

func (s IntervalSchedule) Interval() time.Duration {
	return s.interval
}

func (s IntervalSchedule) String() string {
	return fmt.Sprintf("interval every %s", s.interval)
}
