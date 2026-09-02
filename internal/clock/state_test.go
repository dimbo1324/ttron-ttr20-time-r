package clock

import (
	"testing"
	"time"
)

func TestThresholdsNormalize(t *testing.T) {
	tests := []struct {
		name string
		in   Thresholds
		want Thresholds
	}{
		{
			name: "zero falls back to defaults",
			in:   Thresholds{},
			want: Thresholds{Warn: DefaultWarnThreshold, Critical: DefaultCriticalThreshold},
		},
		{
			name: "negative falls back to defaults",
			in:   Thresholds{Warn: -time.Second, Critical: -time.Minute},
			want: Thresholds{Warn: DefaultWarnThreshold, Critical: DefaultCriticalThreshold},
		},
		{
			name: "critical below warn is raised",
			in:   Thresholds{Warn: 10 * time.Second, Critical: time.Second},
			want: Thresholds{Warn: 10 * time.Second, Critical: 10 * time.Second},
		},
		{
			name: "valid pair is preserved",
			in:   Thresholds{Warn: time.Second, Critical: time.Minute},
			want: Thresholds{Warn: time.Second, Critical: time.Minute},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.Normalize(); got != tt.want {
				t.Fatalf("Normalize() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestThresholdsValidate(t *testing.T) {
	tests := []struct {
		name    string
		in      Thresholds
		wantErr bool
	}{
		{name: "valid", in: Thresholds{Warn: time.Second, Critical: time.Minute}},
		{name: "zero warn", in: Thresholds{Critical: time.Minute}, wantErr: true},
		{name: "zero critical", in: Thresholds{Warn: time.Second}, wantErr: true},
		{name: "critical below warn", in: Thresholds{Warn: time.Minute, Critical: time.Second}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.in.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %t", err, tt.wantErr)
			}
		})
	}
}

func TestThresholdsClassify(t *testing.T) {
	thresholds := Thresholds{Warn: 2 * time.Second, Critical: 30 * time.Second}
	tests := []struct {
		name string
		skew time.Duration
		want State
	}{
		{name: "no skew", skew: 0, want: StateOK},
		{name: "below warn", skew: 1900 * time.Millisecond, want: StateOK},
		{name: "exactly warn", skew: 2 * time.Second, want: StateWarn},
		{name: "negative warn", skew: -2 * time.Second, want: StateWarn},
		{name: "between warn and critical", skew: 10 * time.Second, want: StateWarn},
		{name: "exactly critical", skew: 30 * time.Second, want: StateCritical},
		{name: "negative critical", skew: -45 * time.Second, want: StateCritical},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := thresholds.Classify(tt.skew); got != tt.want {
				t.Fatalf("Classify(%s) = %s, want %s", tt.skew, got, tt.want)
			}
		})
	}
}

func TestStateDegraded(t *testing.T) {
	tests := []struct {
		state State
		want  bool
	}{
		{state: StateUnknown, want: false},
		{state: StateOK, want: false},
		{state: StateWarn, want: true},
		{state: StateCritical, want: true},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			if got := tt.state.Degraded(); got != tt.want {
				t.Fatalf("Degraded() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestDefaultThresholds(t *testing.T) {
	got := DefaultThresholds()
	if got.Warn != DefaultWarnThreshold || got.Critical != DefaultCriticalThreshold {
		t.Fatalf("DefaultThresholds() = %+v", got)
	}
	if err := got.Validate(); err != nil {
		t.Fatalf("DefaultThresholds() must be valid: %v", err)
	}
}
