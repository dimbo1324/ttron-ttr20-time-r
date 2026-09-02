package clock

import (
	"errors"
	"testing"
	"time"
)

var sampleOrigin = time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)

func TestSampleValidate(t *testing.T) {
	tests := []struct {
		name    string
		sample  Sample
		wantErr error
	}{
		{
			name:   "complete sample",
			sample: Sample{RequestedAt: sampleOrigin, ReceivedAt: sampleOrigin.Add(time.Second), DeviceTime: sampleOrigin},
		},
		{
			name:    "missing request time",
			sample:  Sample{ReceivedAt: sampleOrigin, DeviceTime: sampleOrigin},
			wantErr: ErrInvalidSample,
		},
		{
			name:    "missing receive time",
			sample:  Sample{RequestedAt: sampleOrigin, DeviceTime: sampleOrigin},
			wantErr: ErrInvalidSample,
		},
		{
			name:    "missing device time",
			sample:  Sample{RequestedAt: sampleOrigin, ReceivedAt: sampleOrigin},
			wantErr: ErrInvalidSample,
		},
		{
			name:    "received before requested",
			sample:  Sample{RequestedAt: sampleOrigin.Add(time.Second), ReceivedAt: sampleOrigin, DeviceTime: sampleOrigin},
			wantErr: ErrInvalidSample,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sample.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestSampleRoundTripAndReference(t *testing.T) {
	sample := Sample{
		RequestedAt: sampleOrigin,
		ReceivedAt:  sampleOrigin.Add(100 * time.Millisecond),
		DeviceTime:  sampleOrigin.Add(50 * time.Millisecond),
	}

	if got := sample.RoundTrip(); got != 100*time.Millisecond {
		t.Fatalf("RoundTrip() = %s, want 100ms", got)
	}
	if got := sample.Reference(); !got.Equal(sampleOrigin.Add(50 * time.Millisecond)) {
		t.Fatalf("Reference() = %s, want %s", got, sampleOrigin.Add(50*time.Millisecond))
	}
	if got := sample.Skew(); got != 0 {
		t.Fatalf("Skew() = %s, want 0", got)
	}
}

func TestSampleSkewCompensatesLatency(t *testing.T) {
	tests := []struct {
		name       string
		roundTrip  time.Duration
		deviceLead time.Duration
		wantSkew   time.Duration
	}{
		{name: "device ahead", roundTrip: 200 * time.Millisecond, deviceLead: 5 * time.Second, wantSkew: 5*time.Second - 100*time.Millisecond},
		{name: "device behind", roundTrip: 40 * time.Millisecond, deviceLead: -3 * time.Second, wantSkew: -3*time.Second - 20*time.Millisecond},
		{name: "instant response", roundTrip: 0, deviceLead: time.Second, wantSkew: time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sample := Sample{
				RequestedAt: sampleOrigin,
				ReceivedAt:  sampleOrigin.Add(tt.roundTrip),
				DeviceTime:  sampleOrigin.Add(tt.deviceLead),
			}
			if got := sample.Skew(); got != tt.wantSkew {
				t.Fatalf("Skew() = %s, want %s", got, tt.wantSkew)
			}
		})
	}
}

func TestSampleRoundTripOnZeroTimestamps(t *testing.T) {
	if got := (Sample{}).RoundTrip(); got != 0 {
		t.Fatalf("RoundTrip() = %s, want 0", got)
	}
	reversed := Sample{RequestedAt: sampleOrigin.Add(time.Second), ReceivedAt: sampleOrigin, DeviceTime: sampleOrigin}
	if got := reversed.RoundTrip(); got != 0 {
		t.Fatalf("RoundTrip() = %s, want 0", got)
	}
}

func TestAbsDuration(t *testing.T) {
	if got := absDuration(-3 * time.Second); got != 3*time.Second {
		t.Fatalf("absDuration(-3s) = %s", got)
	}
	if got := absDuration(3 * time.Second); got != 3*time.Second {
		t.Fatalf("absDuration(3s) = %s", got)
	}
}
