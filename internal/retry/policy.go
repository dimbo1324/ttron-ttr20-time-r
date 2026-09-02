package retry

import (
	"fmt"
	"time"
)

const (
	DefaultAttempts = 2
	DefaultDelay    = 200 * time.Millisecond
	MaxAttempts     = 16
)

type Policy struct {
	Attempts int
	Delay    time.Duration
	MaxDelay time.Duration
}

func DefaultPolicy() Policy {
	return Policy{Attempts: DefaultAttempts, Delay: DefaultDelay}
}

func (p Policy) Normalize() Policy {
	if p.Attempts < 0 {
		p.Attempts = 0
	}
	if p.Attempts > MaxAttempts {
		p.Attempts = MaxAttempts
	}
	if p.Delay < 0 {
		p.Delay = 0
	}
	if p.MaxDelay <= 0 {
		p.MaxDelay = p.Delay * (1 << uint(min(p.Attempts, 8)))
	}
	if p.MaxDelay < p.Delay {
		p.MaxDelay = p.Delay
	}
	return p
}

func (p Policy) Validate() error {
	if p.Attempts < 0 {
		return fmt.Errorf("retry attempts must not be negative")
	}
	if p.Attempts > MaxAttempts {
		return fmt.Errorf("retry attempts must not exceed %d", MaxAttempts)
	}
	if p.Delay < 0 {
		return fmt.Errorf("retry delay must not be negative")
	}
	return nil
}

func (p Policy) DelayFor(attempt int) time.Duration {
	if attempt <= 0 || p.Delay <= 0 {
		return 0
	}
	delay := p.Delay
	for i := 1; i < attempt; i++ {
		if delay >= p.MaxDelay/2 {
			return p.MaxDelay
		}
		delay *= 2
	}
	if p.MaxDelay > 0 && delay > p.MaxDelay {
		return p.MaxDelay
	}
	return delay
}
