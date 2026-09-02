package schedule

import (
	"errors"
	"testing"
	"time"
)

func TestNewAlignedValidation(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
		offset   time.Duration
		wantErr  error
	}{
		{name: "valid", interval: time.Minute, offset: 5 * time.Second},
		{name: "zero offset", interval: 5 * time.Second},
		{name: "zero interval", interval: 0, wantErr: ErrInvalidInterval},
		{name: "offset equal to interval", interval: time.Minute, offset: time.Minute, wantErr: ErrInvalidOffset},
		{name: "offset above interval", interval: time.Minute, offset: 2 * time.Minute, wantErr: ErrInvalidOffset},
		{name: "negative offset", interval: time.Minute, offset: -time.Second, wantErr: ErrInvalidOffset},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAligned(tt.interval, tt.offset)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewAligned() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestAlignedScheduleFifthSecondOfEveryMinute(t *testing.T) {
	s, err := NewAligned(time.Minute, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		now  time.Time
		want time.Time
	}{
		{
			name: "before the tick",
			now:  time.Date(2026, 9, 2, 17, 3, 4, 0, time.UTC),
			want: time.Date(2026, 9, 2, 17, 3, 5, 0, time.UTC),
		},
		{
			name: "exactly on the tick rolls to the next minute",
			now:  time.Date(2026, 9, 2, 17, 3, 5, 0, time.UTC),
			want: time.Date(2026, 9, 2, 17, 4, 5, 0, time.UTC),
		},
		{
			name: "after the tick",
			now:  time.Date(2026, 9, 2, 17, 3, 7, 250_000_000, time.UTC),
			want: time.Date(2026, 9, 2, 17, 4, 5, 0, time.UTC),
		},
		{
			name: "across the hour boundary",
			now:  time.Date(2026, 9, 2, 17, 59, 30, 0, time.UTC),
			want: time.Date(2026, 9, 2, 18, 0, 5, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.Next(tt.now)
			if !got.Equal(tt.want) {
				t.Fatalf("Next(%s) = %s, want %s", tt.now, got, tt.want)
			}
			if got.Second() != 5 {
				t.Fatalf("Next(%s) second = %d, want 5", tt.now, got.Second())
			}
		})
	}
}

func TestAlignedScheduleEveryFiveSeconds(t *testing.T) {
	s, err := NewAligned(5*time.Second, 0)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		now  time.Time
		want time.Time
	}{
		{
			now:  time.Date(2026, 9, 2, 17, 3, 7, 250_000_000, time.UTC),
			want: time.Date(2026, 9, 2, 17, 3, 10, 0, time.UTC),
		},
		{
			now:  time.Date(2026, 9, 2, 17, 3, 10, 0, time.UTC),
			want: time.Date(2026, 9, 2, 17, 3, 15, 0, time.UTC),
		},
		{
			now:  time.Date(2026, 9, 2, 17, 3, 59, 999_000_000, time.UTC),
			want: time.Date(2026, 9, 2, 17, 4, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.now.Format(time.RFC3339Nano), func(t *testing.T) {
			got := s.Next(tt.now)
			if !got.Equal(tt.want) {
				t.Fatalf("Next(%s) = %s, want %s", tt.now, got, tt.want)
			}
			if got.Second()%5 != 0 || got.Nanosecond() != 0 {
				t.Fatalf("Next(%s) = %s is not aligned to 5s", tt.now, got)
			}
		})
	}
}

func TestAlignedScheduleIsPhaseStable(t *testing.T) {
	s, err := NewAligned(time.Minute, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 9, 2, 17, 0, 0, 0, time.UTC)
	for i := 0; i < 10; i++ {
		next := s.Next(now)
		if !next.After(now) {
			t.Fatalf("Next() = %s is not after %s", next, now)
		}
		if next.Second() != 5 || next.Nanosecond() != 0 {
			t.Fatalf("tick %d = %s lost its phase", i, next)
		}
		now = next.Add(17 * time.Millisecond)
	}
}

func TestAlignedSchedulePreservesLocation(t *testing.T) {
	zone := time.FixedZone("UTC+3", 3*60*60)
	s, err := NewAligned(time.Minute, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 9, 2, 20, 3, 4, 0, zone)
	next := s.Next(now)
	if next.Location() != zone {
		t.Fatalf("Location() = %s, want %s", next.Location(), zone)
	}
	if next.Second() != 5 {
		t.Fatalf("Next() = %s, want the fifth second", next)
	}
}

func TestAlignedScheduleBeforeUnixEpoch(t *testing.T) {
	s, err := NewAligned(time.Minute, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(1969, 12, 31, 23, 58, 30, 0, time.UTC)
	next := s.Next(now)
	if !next.After(now) {
		t.Fatalf("Next() = %s is not after %s", next, now)
	}
	if next.Second() != 5 {
		t.Fatalf("Next() = %s, want the fifth second", next)
	}
}

func TestAlignedScheduleDescription(t *testing.T) {
	withOffset, err := NewAligned(time.Minute, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if got := withOffset.String(); got != "aligned every 1m0s at offset 5s" {
		t.Fatalf("String() = %q", got)
	}
	if got := withOffset.Offset(); got != 5*time.Second {
		t.Fatalf("Offset() = %s", got)
	}

	withoutOffset, err := NewAligned(5*time.Second, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got := withoutOffset.String(); got != "aligned every 5s" {
		t.Fatalf("String() = %q", got)
	}
	if withoutOffset.Mode() != ModeAligned {
		t.Fatalf("Mode() = %q", withoutOffset.Mode())
	}
}

func TestFloorMod(t *testing.T) {
	tests := []struct {
		value   int64
		modulus int64
		want    int64
	}{
		{value: 7, modulus: 5, want: 2},
		{value: -7, modulus: 5, want: 3},
		{value: 0, modulus: 5, want: 0},
		{value: -5, modulus: 5, want: 0},
	}

	for _, tt := range tests {
		if got := floorMod(tt.value, tt.modulus); got != tt.want {
			t.Fatalf("floorMod(%d, %d) = %d, want %d", tt.value, tt.modulus, got, tt.want)
		}
	}
}
