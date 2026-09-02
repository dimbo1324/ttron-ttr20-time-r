package clock

import (
	"sort"
	"time"
)

type entry struct {
	at        time.Time
	skew      time.Duration
	roundTrip time.Duration
}

type window struct {
	capacity int
	entries  []entry
	next     int
	full     bool
}

func newWindow(capacity int) *window {
	if capacity <= 0 {
		capacity = DefaultWindowSize
	}
	return &window{capacity: capacity, entries: make([]entry, capacity)}
}

func (w *window) add(e entry) {
	w.entries[w.next] = e
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

func (w *window) ordered() []entry {
	size := w.len()
	if size == 0 {
		return nil
	}
	out := make([]entry, 0, size)
	if w.full {
		out = append(out, w.entries[w.next:]...)
		out = append(out, w.entries[:w.next]...)
		return out
	}
	out = append(out, w.entries[:w.next]...)
	return out
}

func (w *window) stats() (median, lowest, highest time.Duration, count int) {
	items := w.ordered()
	count = len(items)
	if count == 0 {
		return 0, 0, 0, 0
	}
	skews := make([]time.Duration, count)
	for i, item := range items {
		skews[i] = item.skew
	}
	lowest, highest = skews[0], skews[0]
	for _, skew := range skews {
		if skew < lowest {
			lowest = skew
		}
		if skew > highest {
			highest = skew
		}
	}
	sort.Slice(skews, func(i, j int) bool { return skews[i] < skews[j] })
	middle := count / 2
	if count%2 == 1 {
		median = skews[middle]
	} else {
		median = (skews[middle-1] + skews[middle]) / 2
	}
	return median, lowest, highest, count
}
