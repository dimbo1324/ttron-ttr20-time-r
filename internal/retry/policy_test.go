package retry

import (
	"testing"
	"time"
)

func TestDefaultPolicyIsValid(t *testing.T) {
	policy := DefaultPolicy()
	if err := policy.Validate(); err != nil {
		t.Fatalf("DefaultPolicy() must be valid: %v", err)
	}
	if policy.Attempts != DefaultAttempts || policy.Delay != DefaultDelay {
		t.Fatalf("DefaultPolicy() = %+v", policy)
	}
}

func TestPolicyNormalize(t *testing.T) {
	tests := []struct {
		name string
		in   Policy
		want Policy
	}{
		{
			name: "negative attempts clamp to zero",
			in:   Policy{Attempts: -3, Delay: 100 * time.Millisecond},
			want: Policy{Attempts: 0, Delay: 100 * time.Millisecond, MaxDelay: 100 * time.Millisecond},
		},
		{
			name: "attempts clamp to the maximum",
			in:   Policy{Attempts: 1000, Delay: time.Millisecond},
			want: Policy{Attempts: MaxAttempts, Delay: time.Millisecond, MaxDelay: 256 * time.Millisecond},
		},
		{
			name: "negative delay clamps to zero",
			in:   Policy{Attempts: 2, Delay: -time.Second},
			want: Policy{Attempts: 2, Delay: 0, MaxDelay: 0},
		},
		{
			name: "max delay is derived from attempts",
			in:   Policy{Attempts: 2, Delay: 200 * time.Millisecond},
			want: Policy{Attempts: 2, Delay: 200 * time.Millisecond, MaxDelay: 800 * time.Millisecond},
		},
		{
			name: "max delay below delay is raised",
			in:   Policy{Attempts: 2, Delay: time.Second, MaxDelay: time.Millisecond},
			want: Policy{Attempts: 2, Delay: time.Second, MaxDelay: time.Second},
		},
		{
			name: "explicit max delay is preserved",
			in:   Policy{Attempts: 3, Delay: 100 * time.Millisecond, MaxDelay: 250 * time.Millisecond},
			want: Policy{Attempts: 3, Delay: 100 * time.Millisecond, MaxDelay: 250 * time.Millisecond},
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

func TestPolicyValidate(t *testing.T) {
	tests := []struct {
		name    string
		in      Policy
		wantErr bool
	}{
		{name: "valid", in: Policy{Attempts: 2, Delay: time.Millisecond}},
		{name: "zero attempts allowed", in: Policy{Attempts: 0, Delay: time.Millisecond}},
		{name: "negative attempts", in: Policy{Attempts: -1}, wantErr: true},
		{name: "too many attempts", in: Policy{Attempts: MaxAttempts + 1}, wantErr: true},
		{name: "negative delay", in: Policy{Attempts: 1, Delay: -time.Second}, wantErr: true},
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

func TestPolicyDelayForGrowsExponentially(t *testing.T) {
	policy := Policy{Attempts: 4, Delay: 100 * time.Millisecond}.Normalize()

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{attempt: 0, want: 0},
		{attempt: -1, want: 0},
		{attempt: 1, want: 100 * time.Millisecond},
		{attempt: 2, want: 200 * time.Millisecond},
		{attempt: 3, want: 400 * time.Millisecond},
		{attempt: 4, want: 800 * time.Millisecond},
		{attempt: 5, want: policy.MaxDelay},
		{attempt: 50, want: policy.MaxDelay},
	}

	for _, tt := range tests {
		if got := policy.DelayFor(tt.attempt); got != tt.want {
			t.Fatalf("DelayFor(%d) = %s, want %s", tt.attempt, got, tt.want)
		}
	}
}

func TestPolicyDelayForNeverExceedsMaxDelay(t *testing.T) {
	policy := Policy{Attempts: 8, Delay: 50 * time.Millisecond, MaxDelay: 120 * time.Millisecond}.Normalize()

	for attempt := 1; attempt <= 12; attempt++ {
		if got := policy.DelayFor(attempt); got > policy.MaxDelay {
			t.Fatalf("DelayFor(%d) = %s exceeds MaxDelay %s", attempt, got, policy.MaxDelay)
		}
	}
}

func TestPolicyDelayForZeroDelay(t *testing.T) {
	policy := Policy{Attempts: 3}.Normalize()
	for attempt := 0; attempt <= 4; attempt++ {
		if got := policy.DelayFor(attempt); got != 0 {
			t.Fatalf("DelayFor(%d) = %s, want 0", attempt, got)
		}
	}
}
