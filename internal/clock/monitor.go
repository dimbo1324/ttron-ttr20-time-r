package clock

import (
	"sync"
	"time"
)

type Report struct {
	At            time.Time
	DeviceTime    time.Time
	Reference     time.Time
	Skew          time.Duration
	AbsSkew       time.Duration
	MedianSkew    time.Duration
	MinSkew       time.Duration
	MaxSkew       time.Duration
	RoundTrip     time.Duration
	Drift         Drift
	State         State
	PreviousState State
	Transitioned  bool
	SampleCount   int
	Thresholds    Thresholds
}

func (r Report) Observed() bool {
	return !r.At.IsZero()
}

type Monitor struct {
	mu sync.RWMutex

	thresholds Thresholds
	window     *window
	state      State
	last       Report

	observed uint64
	rejected uint64
}

func NewMonitor(thresholds Thresholds, windowSize int) *Monitor {
	return &Monitor{
		thresholds: thresholds.Normalize(),
		window:     newWindow(windowSize),
		state:      StateUnknown,
	}
}

func (m *Monitor) Observe(sample Sample) (Report, error) {
	if err := sample.Validate(); err != nil {
		m.mu.Lock()
		m.rejected++
		m.mu.Unlock()
		return Report{}, err
	}

	skew := sample.Skew()
	roundTrip := sample.RoundTrip()
	reference := sample.Reference()

	m.mu.Lock()
	defer m.mu.Unlock()

	m.observed++
	m.window.add(entry{at: reference, skew: skew, roundTrip: roundTrip})

	median, min, max, count := m.window.stats()
	drift := computeDrift(m.window.ordered())

	classified := skew
	if count >= MinDriftSamples {
		classified = median
	}

	previous := m.state
	next := m.thresholds.Classify(classified)
	m.state = next

	report := Report{
		At:            sample.ReceivedAt,
		DeviceTime:    sample.DeviceTime,
		Reference:     reference,
		Skew:          skew,
		AbsSkew:       absDuration(skew),
		MedianSkew:    median,
		MinSkew:       min,
		MaxSkew:       max,
		RoundTrip:     roundTrip,
		Drift:         drift,
		State:         next,
		PreviousState: previous,
		Transitioned:  previous != next,
		SampleCount:   count,
		Thresholds:    m.thresholds,
	}
	m.last = report
	return report, nil
}

func (m *Monitor) Report() Report {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.last
}

func (m *Monitor) State() State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

func (m *Monitor) Thresholds() Thresholds {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.thresholds
}

func (m *Monitor) SetThresholds(thresholds Thresholds) Thresholds {
	normalized := thresholds.Normalize()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.thresholds = normalized
	return normalized
}

func (m *Monitor) Counters() (observed, rejected uint64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.observed, m.rejected
}

func (m *Monitor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.window.reset()
	m.state = StateUnknown
	m.last = Report{}
}

// Point is one skew measurement, at the reference instant it belongs to --
// the request time plus half the round trip, which is where the device most
// plausibly read its own clock.
type Point struct {
	At        time.Time
	Skew      time.Duration
	RoundTrip time.Duration
}

// Samples returns the skew window, oldest first, so a caller can draw the
// history the median and the drift line were computed from.
func (m *Monitor) Samples() []Point {
	m.mu.RLock()
	defer m.mu.RUnlock()

	items := m.window.ordered()
	out := make([]Point, 0, len(items))
	for _, item := range items {
		out = append(out, Point{At: item.at, Skew: item.skew, RoundTrip: item.roundTrip})
	}
	return out
}
