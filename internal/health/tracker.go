package health

import (
	"sync"
	"time"
)

type Transition struct {
	From    State
	To      State
	Changed bool
	At      time.Time
	Reason  string
}

type Snapshot struct {
	State                State
	Since                time.Time
	Policy               Policy
	Availability         float64
	WindowSamples        int
	ConsecutiveFailures  int
	ConsecutiveSuccesses int
	TotalSuccesses       uint64
	TotalFailures        uint64
	LastSuccessAt        time.Time
	LastFailureAt        time.Time
	LastError            string
	Latency              Latency
}

type Tracker struct {
	mu sync.RWMutex

	policy Policy
	window *window
	now    func() time.Time

	state State
	since time.Time

	consecutiveFailures  int
	consecutiveSuccesses int
	totalSuccesses       uint64
	totalFailures        uint64
	lastSuccessAt        time.Time
	lastFailureAt        time.Time
	lastError            string
}

func NewTracker(policy Policy) *Tracker {
	normalized := policy.Normalize()
	return &Tracker{
		policy: normalized,
		window: newWindow(normalized.WindowSize),
		now:    time.Now,
		state:  StateUnknown,
	}
}

func (t *Tracker) WithClock(now func() time.Time) *Tracker {
	if now != nil {
		t.mu.Lock()
		t.now = now
		t.mu.Unlock()
	}
	return t
}

func (t *Tracker) RecordSuccess(latency time.Duration) Transition {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	t.window.add(outcome{at: now, success: true, latency: latency})
	t.totalSuccesses++
	t.consecutiveFailures = 0
	t.consecutiveSuccesses++
	t.lastSuccessAt = now
	t.lastError = ""

	return t.transition(t.stateAfterSuccess(), now, "successful read")
}

func (t *Tracker) RecordFailure(err error) Transition {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	t.window.add(outcome{at: now, success: false})
	t.totalFailures++
	t.consecutiveSuccesses = 0
	t.consecutiveFailures++
	t.lastFailureAt = now
	reason := "failed read"
	if err != nil {
		t.lastError = err.Error()
		reason = err.Error()
	}

	return t.transition(t.stateAfterFailure(), now, reason)
}

func (t *Tracker) stateAfterSuccess() State {
	switch t.state {
	case StateOnline:
		return StateOnline
	case StateUnknown:
		return StateOnline
	default:
		if t.consecutiveSuccesses >= t.policy.RecoverAfter {
			return StateOnline
		}
		return t.state
	}
}

func (t *Tracker) stateAfterFailure() State {
	switch {
	case t.consecutiveFailures >= t.policy.OfflineAfter:
		return StateOffline
	case t.consecutiveFailures >= t.policy.DegradeAfter:
		return StateDegraded
	case t.state == StateUnknown:
		return StateUnknown
	default:
		return t.state
	}
}

func (t *Tracker) transition(next State, at time.Time, reason string) Transition {
	previous := t.state
	if previous == next {
		return Transition{From: previous, To: next, At: at, Reason: reason}
	}
	t.state = next
	t.since = at
	return Transition{From: previous, To: next, Changed: true, At: at, Reason: reason}
}

func (t *Tracker) State() State {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state
}

func (t *Tracker) Snapshot() Snapshot {
	t.mu.RLock()
	defer t.mu.RUnlock()

	availability, samples := t.window.availability()
	return Snapshot{
		State:                t.state,
		Since:                t.since,
		Policy:               t.policy,
		Availability:         availability,
		WindowSamples:        samples,
		ConsecutiveFailures:  t.consecutiveFailures,
		ConsecutiveSuccesses: t.consecutiveSuccesses,
		TotalSuccesses:       t.totalSuccesses,
		TotalFailures:        t.totalFailures,
		LastSuccessAt:        t.lastSuccessAt,
		LastFailureAt:        t.lastFailureAt,
		LastError:            t.lastError,
		Latency:              t.window.latencies(),
	}
}

func (t *Tracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.window.reset()
	t.state = StateUnknown
	t.since = time.Time{}
	t.consecutiveFailures = 0
	t.consecutiveSuccesses = 0
	t.lastError = ""
}
