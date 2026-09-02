package clock

import (
	"testing"
	"time"
)

func makeEntries(skews ...time.Duration) []entry {
	out := make([]entry, 0, len(skews))
	for i, skew := range skews {
		out = append(out, entry{at: sampleOrigin.Add(time.Duration(i) * time.Second), skew: skew})
	}
	return out
}

func TestWindowCapacityFallback(t *testing.T) {
	w := newWindow(0)
	if w.capacity != DefaultWindowSize {
		t.Fatalf("capacity = %d, want %d", w.capacity, DefaultWindowSize)
	}
	if w := newWindow(-5); w.capacity != DefaultWindowSize {
		t.Fatalf("capacity = %d, want %d", w.capacity, DefaultWindowSize)
	}
}

func TestWindowOverwritesOldestEntries(t *testing.T) {
	w := newWindow(3)
	for _, e := range makeEntries(time.Second, 2*time.Second, 3*time.Second, 4*time.Second) {
		w.add(e)
	}

	if got := w.len(); got != 3 {
		t.Fatalf("len() = %d, want 3", got)
	}
	ordered := w.ordered()
	want := []time.Duration{2 * time.Second, 3 * time.Second, 4 * time.Second}
	for i, item := range ordered {
		if item.skew != want[i] {
			t.Fatalf("ordered()[%d].skew = %s, want %s", i, item.skew, want[i])
		}
	}
}

func TestWindowOrderedPartial(t *testing.T) {
	w := newWindow(4)
	for _, e := range makeEntries(time.Second, 2*time.Second) {
		w.add(e)
	}
	ordered := w.ordered()
	if len(ordered) != 2 {
		t.Fatalf("ordered() length = %d, want 2", len(ordered))
	}
	if ordered[0].skew != time.Second || ordered[1].skew != 2*time.Second {
		t.Fatalf("ordered() = %+v", ordered)
	}
}

func TestWindowEmpty(t *testing.T) {
	w := newWindow(3)
	if got := w.ordered(); got != nil {
		t.Fatalf("ordered() = %+v, want nil", got)
	}
	median, lowest, highest, count := w.stats()
	if median != 0 || lowest != 0 || highest != 0 || count != 0 {
		t.Fatalf("stats() = (%s, %s, %s, %d), want zeros", median, lowest, highest, count)
	}
}

func TestWindowStats(t *testing.T) {
	tests := []struct {
		name        string
		skews       []time.Duration
		wantMedian  time.Duration
		wantLowest  time.Duration
		wantHighest time.Duration
	}{
		{
			name:        "odd count",
			skews:       []time.Duration{3 * time.Second, time.Second, 2 * time.Second},
			wantMedian:  2 * time.Second,
			wantLowest:  time.Second,
			wantHighest: 3 * time.Second,
		},
		{
			name:        "even count averages middle pair",
			skews:       []time.Duration{4 * time.Second, 2 * time.Second, 6 * time.Second, 8 * time.Second},
			wantMedian:  5 * time.Second,
			wantLowest:  2 * time.Second,
			wantHighest: 8 * time.Second,
		},
		{
			name:        "negative skews",
			skews:       []time.Duration{-5 * time.Second, -time.Second, -3 * time.Second},
			wantMedian:  -3 * time.Second,
			wantLowest:  -5 * time.Second,
			wantHighest: -time.Second,
		},
		{
			name:        "single sample",
			skews:       []time.Duration{7 * time.Second},
			wantMedian:  7 * time.Second,
			wantLowest:  7 * time.Second,
			wantHighest: 7 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWindow(16)
			for _, e := range makeEntries(tt.skews...) {
				w.add(e)
			}
			median, lowest, highest, count := w.stats()
			if count != len(tt.skews) {
				t.Fatalf("count = %d, want %d", count, len(tt.skews))
			}
			if median != tt.wantMedian {
				t.Fatalf("median = %s, want %s", median, tt.wantMedian)
			}
			if lowest != tt.wantLowest || highest != tt.wantHighest {
				t.Fatalf("range = (%s, %s), want (%s, %s)", lowest, highest, tt.wantLowest, tt.wantHighest)
			}
		})
	}
}

func TestWindowReset(t *testing.T) {
	w := newWindow(2)
	for _, e := range makeEntries(time.Second, 2*time.Second, 3*time.Second) {
		w.add(e)
	}
	w.reset()
	if w.len() != 0 || w.full {
		t.Fatalf("after reset len = %d, full = %t", w.len(), w.full)
	}
}
