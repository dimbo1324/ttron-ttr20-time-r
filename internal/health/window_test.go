package health

import (
	"testing"
	"time"
)

func TestWindowCapacityFallback(t *testing.T) {
	for _, capacity := range []int{0, -1} {
		if got := newWindow(capacity).capacity; got != DefaultWindowSize {
			t.Fatalf("newWindow(%d).capacity = %d, want %d", capacity, got, DefaultWindowSize)
		}
	}
}

func TestWindowAvailability(t *testing.T) {
	tests := []struct {
		name        string
		outcomes    []bool
		wantRatio   float64
		wantSamples int
	}{
		{name: "empty", wantRatio: 0, wantSamples: 0},
		{name: "all success", outcomes: []bool{true, true, true}, wantRatio: 1, wantSamples: 3},
		{name: "all failure", outcomes: []bool{false, false}, wantRatio: 0, wantSamples: 2},
		{name: "half", outcomes: []bool{true, false, true, false}, wantRatio: 0.5, wantSamples: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWindow(8)
			for _, success := range tt.outcomes {
				w.add(outcome{at: time.Now(), success: success, latency: time.Millisecond})
			}
			ratio, samples := w.availability()
			if ratio != tt.wantRatio || samples != tt.wantSamples {
				t.Fatalf("availability() = (%f, %d), want (%f, %d)", ratio, samples, tt.wantRatio, tt.wantSamples)
			}
		})
	}
}

func TestWindowKeepsOnlyRecentOutcomes(t *testing.T) {
	w := newWindow(3)
	for i := 0; i < 5; i++ {
		w.add(outcome{at: time.Now(), success: i >= 2, latency: time.Millisecond})
	}

	ratio, samples := w.availability()
	if samples != 3 {
		t.Fatalf("samples = %d, want 3", samples)
	}
	if ratio != 1 {
		t.Fatalf("availability = %f, want 1", ratio)
	}
	if got := w.len(); got != 3 {
		t.Fatalf("len() = %d, want 3", got)
	}
}

func TestWindowLatenciesIgnoreFailures(t *testing.T) {
	w := newWindow(8)
	w.add(outcome{success: true, latency: 10 * time.Millisecond})
	w.add(outcome{success: false, latency: 900 * time.Millisecond})
	w.add(outcome{success: true, latency: 30 * time.Millisecond})
	w.add(outcome{success: true, latency: 20 * time.Millisecond})

	latency := w.latencies()
	if latency.Samples != 3 {
		t.Fatalf("Samples = %d, want 3", latency.Samples)
	}
	if latency.Min != 10*time.Millisecond || latency.Max != 30*time.Millisecond {
		t.Fatalf("range = (%s, %s)", latency.Min, latency.Max)
	}
	if latency.Mean != 20*time.Millisecond {
		t.Fatalf("Mean = %s, want 20ms", latency.Mean)
	}
	if latency.P50 != 20*time.Millisecond {
		t.Fatalf("P50 = %s, want 20ms", latency.P50)
	}
}

func TestWindowLatenciesEmptyWithoutSuccesses(t *testing.T) {
	w := newWindow(4)
	w.add(outcome{success: false})
	w.add(outcome{success: false})

	if got := w.latencies(); got != (Latency{}) {
		t.Fatalf("latencies() = %+v, want zero", got)
	}
}

func TestWindowLatenciesKeepSubResolutionSamples(t *testing.T) {
	w := newWindow(4)
	w.add(outcome{success: true, latency: 0})
	w.add(outcome{success: true, latency: 2 * time.Millisecond})

	latency := w.latencies()
	if latency.Samples != 2 {
		t.Fatalf("Samples = %d, want 2: a zero measurement is still a measurement", latency.Samples)
	}
	if latency.Min != 0 || latency.Max != 2*time.Millisecond {
		t.Fatalf("range = (%s, %s)", latency.Min, latency.Max)
	}
	if latency.Mean != time.Millisecond {
		t.Fatalf("Mean = %s, want 1ms", latency.Mean)
	}
}

func TestPercentile(t *testing.T) {
	sorted := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
	}

	tests := []struct {
		name   string
		values []time.Duration
		q      float64
		want   time.Duration
	}{
		{name: "empty", values: nil, q: 0.5, want: 0},
		{name: "single", values: sorted[:1], q: 0.99, want: 10 * time.Millisecond},
		{name: "median", values: sorted, q: 0.5, want: 30 * time.Millisecond},
		{name: "p95", values: sorted, q: 0.95, want: 50 * time.Millisecond},
		{name: "p99", values: sorted, q: 0.99, want: 50 * time.Millisecond},
		{name: "min", values: sorted, q: 0, want: 10 * time.Millisecond},
		{name: "clamped above", values: sorted, q: 5, want: 50 * time.Millisecond},
		{name: "clamped below", values: sorted, q: -5, want: 10 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := percentile(tt.values, tt.q); got != tt.want {
				t.Fatalf("percentile(%f) = %s, want %s", tt.q, got, tt.want)
			}
		})
	}
}

func TestWindowReset(t *testing.T) {
	w := newWindow(2)
	w.add(outcome{success: true, latency: time.Millisecond})
	w.add(outcome{success: true, latency: time.Millisecond})
	w.add(outcome{success: true, latency: time.Millisecond})

	w.reset()
	if w.len() != 0 || w.ordered() != nil {
		t.Fatalf("after reset len = %d", w.len())
	}
}
