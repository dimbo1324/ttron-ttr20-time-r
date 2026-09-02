package schedule

import (
	"errors"
	"testing"
	"time"
)

func TestParseMode(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    Mode
		wantErr error
	}{
		{name: "empty defaults to interval", in: "", want: ModeInterval},
		{name: "interval", in: "interval", want: ModeInterval},
		{name: "aligned", in: "aligned", want: ModeAligned},
		{name: "upper case", in: "ALIGNED", want: ModeAligned},
		{name: "padded", in: "  aligned  ", want: ModeAligned},
		{name: "unknown", in: "cron", wantErr: ErrInvalidMode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMode(tt.in)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ParseMode(%q) error = %v, want %v", tt.in, err, tt.wantErr)
			}
			if tt.wantErr == nil && got != tt.want {
				t.Fatalf("ParseMode(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestNewSchedule(t *testing.T) {
	tests := []struct {
		name     string
		mode     Mode
		interval time.Duration
		offset   time.Duration
		wantMode Mode
		wantErr  error
	}{
		{name: "interval", mode: ModeInterval, interval: 5 * time.Second, wantMode: ModeInterval},
		{name: "empty mode", mode: "", interval: 5 * time.Second, wantMode: ModeInterval},
		{name: "aligned", mode: ModeAligned, interval: time.Minute, offset: 5 * time.Second, wantMode: ModeAligned},
		{name: "zero interval", mode: ModeInterval, interval: 0, wantErr: ErrInvalidInterval},
		{name: "negative interval", mode: ModeAligned, interval: -time.Second, wantErr: ErrInvalidInterval},
		{name: "offset beyond interval", mode: ModeAligned, interval: time.Second, offset: 2 * time.Second, wantErr: ErrInvalidOffset},
		{name: "negative offset", mode: ModeAligned, interval: time.Minute, offset: -time.Second, wantErr: ErrInvalidOffset},
		{name: "unknown mode", mode: Mode("cron"), interval: time.Second, wantErr: ErrInvalidMode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.mode, tt.interval, tt.offset)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("New() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if got.Mode() != tt.wantMode {
				t.Fatalf("Mode() = %q, want %q", got.Mode(), tt.wantMode)
			}
			if got.Interval() != tt.interval {
				t.Fatalf("Interval() = %s, want %s", got.Interval(), tt.interval)
			}
			if got.String() == "" {
				t.Fatal("String() must not be empty")
			}
		})
	}
}
