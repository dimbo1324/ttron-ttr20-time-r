package schedule

import (
	"errors"
	"testing"
	"time"
)

func TestNewIntervalRejectsNonPositive(t *testing.T) {
	for _, interval := range []time.Duration{0, -time.Second} {
		if _, err := NewInterval(interval); !errors.Is(err, ErrInvalidInterval) {
			t.Fatalf("NewInterval(%s) error = %v, want %v", interval, err, ErrInvalidInterval)
		}
	}
}

func TestIntervalScheduleNext(t *testing.T) {
	s, err := NewInterval(5 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 9, 2, 17, 3, 7, 250_000_000, time.UTC)
	want := now.Add(5 * time.Second)
	if got := s.Next(now); !got.Equal(want) {
		t.Fatalf("Next() = %s, want %s", got, want)
	}
	if s.Mode() != ModeInterval {
		t.Fatalf("Mode() = %q", s.Mode())
	}
	if s.Interval() != 5*time.Second {
		t.Fatalf("Interval() = %s", s.Interval())
	}
	if s.String() != "interval every 5s" {
		t.Fatalf("String() = %q", s.String())
	}
}

func TestIntervalScheduleIsMonotonic(t *testing.T) {
	s, err := NewInterval(time.Second)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 2, 17, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		next := s.Next(now)
		if !next.After(now) {
			t.Fatalf("Next() = %s is not after %s", next, now)
		}
		now = next
	}
}
