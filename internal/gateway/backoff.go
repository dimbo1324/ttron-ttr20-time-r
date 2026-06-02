package gateway

import "time"

type Backoff struct {
	initial time.Duration
	max     time.Duration
	current time.Duration
}

func NewBackoff(initial, max time.Duration) Backoff {
	if initial <= 0 {
		initial = 500 * time.Millisecond
	}
	if max <= 0 {
		max = 5 * time.Second
	}
	if max < initial {
		max = initial
	}
	return Backoff{initial: initial, max: max, current: initial}
}

func (b *Backoff) Next() time.Duration {
	out := b.current
	b.current *= 2
	if b.current > b.max {
		b.current = b.max
	}
	return out
}

func (b *Backoff) Reset() {
	b.current = b.initial
}
