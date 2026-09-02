package schedule

import (
	"context"
	"time"
)

type Ticker struct {
	schedule Schedule
	now      func() time.Time
}

func NewTicker(s Schedule) *Ticker {
	return &Ticker{schedule: s, now: time.Now}
}

func (t *Ticker) WithClock(now func() time.Time) *Ticker {
	if now != nil {
		t.now = now
	}
	return t
}

func (t *Ticker) NextAt() time.Time {
	return t.schedule.Next(t.now())
}

func (t *Ticker) Wait(ctx context.Context) (time.Time, bool) {
	return t.WaitUntil(ctx, t.schedule.Next(t.now()))
}

func (t *Ticker) WaitUntil(ctx context.Context, next time.Time) (time.Time, bool) {
	delay := next.Sub(t.now())
	if delay <= 0 {
		select {
		case <-ctx.Done():
			return time.Time{}, false
		default:
			return next, true
		}
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return time.Time{}, false
	case tick := <-timer.C:
		return tick, true
	}
}

func (t *Ticker) Schedule() Schedule {
	return t.schedule
}
