package health

import (
	"sort"
	"time"
)

type outcome struct {
	at      time.Time
	success bool
	latency time.Duration
}

type window struct {
	capacity int
	items    []outcome
	next     int
	full     bool
}

func newWindow(capacity int) *window {
	if capacity <= 0 {
		capacity = DefaultWindowSize
	}
	return &window{capacity: capacity, items: make([]outcome, capacity)}
}

func (w *window) add(o outcome) {
	w.items[w.next] = o
	w.next = (w.next + 1) % w.capacity
	if w.next == 0 {
		w.full = true
	}
}

func (w *window) len() int {
	if w.full {
		return w.capacity
	}
	return w.next
}

func (w *window) reset() {
	w.next = 0
	w.full = false
}

func (w *window) ordered() []outcome {
	size := w.len()
	if size == 0 {
		return nil
	}
	out := make([]outcome, 0, size)
	if w.full {
		out = append(out, w.items[w.next:]...)
		out = append(out, w.items[:w.next]...)
		return out
	}
	return append(out, w.items[:w.next]...)
}

func (w *window) availability() (ratio float64, samples int) {
	items := w.ordered()
	if len(items) == 0 {
		return 0, 0
	}
	success := 0
	for _, item := range items {
		if item.success {
			success++
		}
	}
	return float64(success) / float64(len(items)), len(items)
}

func (w *window) latencies() Latency {
	items := w.ordered()
	values := make([]time.Duration, 0, len(items))
	for _, item := range items {
		if item.success && item.latency > 0 {
			values = append(values, item.latency)
		}
	}
	if len(values) == 0 {
		return Latency{}
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })

	var total time.Duration
	for _, value := range values {
		total += value
	}
	return Latency{
		Samples: len(values),
		Min:     values[0],
		Max:     values[len(values)-1],
		Mean:    total / time.Duration(len(values)),
		P50:     percentile(values, 0.50),
		P95:     percentile(values, 0.95),
		P99:     percentile(values, 0.99),
	}
}

type Latency struct {
	Samples int
	Min     time.Duration
	Max     time.Duration
	Mean    time.Duration
	P50     time.Duration
	P95     time.Duration
	P99     time.Duration
}

func percentile(sorted []time.Duration, q float64) time.Duration {
	count := len(sorted)
	if count == 0 {
		return 0
	}
	if count == 1 {
		return sorted[0]
	}
	index := int(float64(count-1)*q + 0.5)
	if index < 0 {
		index = 0
	}
	if index >= count {
		index = count - 1
	}
	return sorted[index]
}
