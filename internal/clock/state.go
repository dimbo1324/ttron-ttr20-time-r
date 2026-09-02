package clock

import (
	"fmt"
	"time"
)

type State string

const (
	StateUnknown  State = "unknown"
	StateOK       State = "ok"
	StateWarn     State = "warn"
	StateCritical State = "critical"
)

const (
	DefaultWarnThreshold     = 2 * time.Second
	DefaultCriticalThreshold = 30 * time.Second
	DefaultWindowSize        = 64
)

type Thresholds struct {
	Warn     time.Duration
	Critical time.Duration
}

func DefaultThresholds() Thresholds {
	return Thresholds{Warn: DefaultWarnThreshold, Critical: DefaultCriticalThreshold}
}

func (t Thresholds) Normalize() Thresholds {
	if t.Warn <= 0 {
		t.Warn = DefaultWarnThreshold
	}
	if t.Critical <= 0 {
		t.Critical = DefaultCriticalThreshold
	}
	if t.Critical < t.Warn {
		t.Critical = t.Warn
	}
	return t
}

func (t Thresholds) Validate() error {
	if t.Warn <= 0 {
		return fmt.Errorf("clock warn threshold must be positive")
	}
	if t.Critical <= 0 {
		return fmt.Errorf("clock critical threshold must be positive")
	}
	if t.Critical < t.Warn {
		return fmt.Errorf("clock critical threshold must not be below warn threshold")
	}
	return nil
}

func (t Thresholds) Classify(skew time.Duration) State {
	abs := absDuration(skew)
	switch {
	case abs >= t.Critical:
		return StateCritical
	case abs >= t.Warn:
		return StateWarn
	default:
		return StateOK
	}
}

func (s State) Degraded() bool {
	return s == StateWarn || s == StateCritical
}
