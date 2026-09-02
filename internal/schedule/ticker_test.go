package schedule

import (
	"context"
	"sync"
	"testing"
	"time"
)

type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func TestTickerNextAtUsesInjectedClock(t *testing.T) {
	s, err := NewAligned(time.Minute, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	clock := &fakeClock{now: time.Date(2026, 9, 2, 17, 3, 4, 0, time.UTC)}
	ticker := NewTicker(s).WithClock(clock.Now)

	want := time.Date(2026, 9, 2, 17, 3, 5, 0, time.UTC)
	if got := ticker.NextAt(); !got.Equal(want) {
		t.Fatalf("NextAt() = %s, want %s", got, want)
	}
	if ticker.Schedule() != s {
		t.Fatal("Schedule() must return the configured schedule")
	}
}

func TestTickerWithNilClockKeepsDefault(t *testing.T) {
	s, err := NewInterval(time.Second)
	if err != nil {
		t.Fatal(err)
	}
	ticker := NewTicker(s).WithClock(nil)
	if ticker.now == nil {
		t.Fatal("clock must not be nil")
	}
}

func TestTickerWaitFires(t *testing.T) {
	s, err := NewInterval(20 * time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	ticker := NewTicker(s)

	start := time.Now()
	tick, ok := ticker.Wait(context.Background())
	if !ok {
		t.Fatal("Wait() must fire")
	}
	if elapsed := time.Since(start); elapsed < 10*time.Millisecond {
		t.Fatalf("Wait() returned after %s, want at least 10ms", elapsed)
	}
	if tick.IsZero() {
		t.Fatal("Wait() must return the tick time")
	}
}

func TestTickerWaitStopsOnCancelledContext(t *testing.T) {
	s, err := NewInterval(time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	ticker := NewTicker(s)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	tick, ok := ticker.Wait(ctx)
	if ok {
		t.Fatal("Wait() must report cancellation")
	}
	if !tick.IsZero() {
		t.Fatalf("Wait() = %s, want zero time on cancellation", tick)
	}
}

func TestTickerWaitUntilPastDeadlineReturnsImmediately(t *testing.T) {
	s, err := NewInterval(time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	ticker := NewTicker(s)

	past := time.Now().Add(-time.Minute)
	start := time.Now()
	tick, ok := ticker.WaitUntil(context.Background(), past)
	if !ok {
		t.Fatal("WaitUntil() must fire for a past deadline")
	}
	if !tick.Equal(past) {
		t.Fatalf("WaitUntil() = %s, want %s", tick, past)
	}
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("WaitUntil() blocked for %s", elapsed)
	}
}

func TestTickerWaitUntilPastDeadlineRespectsCancellation(t *testing.T) {
	s, err := NewInterval(time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	ticker := NewTicker(s)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, ok := ticker.WaitUntil(ctx, time.Now().Add(-time.Minute)); ok {
		t.Fatal("WaitUntil() must not fire when the context is already done")
	}
}
