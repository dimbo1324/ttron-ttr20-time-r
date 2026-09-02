package clock

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func sampleWithSkew(at time.Time, skew time.Duration) Sample {
	return Sample{RequestedAt: at, ReceivedAt: at, DeviceTime: at.Add(skew)}
}

func testMonitor() *Monitor {
	return NewMonitor(Thresholds{Warn: 2 * time.Second, Critical: 30 * time.Second}, 16)
}

func TestMonitorRejectsInvalidSample(t *testing.T) {
	monitor := testMonitor()

	report, err := monitor.Observe(Sample{})
	if !errors.Is(err, ErrInvalidSample) {
		t.Fatalf("Observe() error = %v, want %v", err, ErrInvalidSample)
	}
	if report.Observed() {
		t.Fatalf("Observe() returned a report for an invalid sample: %+v", report)
	}
	if monitor.State() != StateUnknown {
		t.Fatalf("State() = %s, want %s", monitor.State(), StateUnknown)
	}

	observed, rejected := monitor.Counters()
	if observed != 0 || rejected != 1 {
		t.Fatalf("Counters() = (%d, %d), want (0, 1)", observed, rejected)
	}
}

func TestMonitorFirstObservationUsesLastSkew(t *testing.T) {
	monitor := testMonitor()

	report, err := monitor.Observe(sampleWithSkew(sampleOrigin, 5*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if report.State != StateWarn {
		t.Fatalf("State = %s, want %s", report.State, StateWarn)
	}
	if report.PreviousState != StateUnknown || !report.Transitioned {
		t.Fatalf("transition = %s -> %s (changed=%t)", report.PreviousState, report.State, report.Transitioned)
	}
	if report.Skew != 5*time.Second || report.AbsSkew != 5*time.Second {
		t.Fatalf("Skew = %s, AbsSkew = %s", report.Skew, report.AbsSkew)
	}
	if report.SampleCount != 1 {
		t.Fatalf("SampleCount = %d, want 1", report.SampleCount)
	}
	if !report.Observed() {
		t.Fatal("report must be marked as observed")
	}
}

func TestMonitorClassifiesOnMedianAfterEnoughSamples(t *testing.T) {
	monitor := testMonitor()

	for i := 0; i < MinDriftSamples; i++ {
		if _, err := monitor.Observe(sampleWithSkew(sampleOrigin.Add(time.Duration(i)*time.Minute), 0)); err != nil {
			t.Fatal(err)
		}
	}
	if monitor.State() != StateOK {
		t.Fatalf("State() = %s, want %s", monitor.State(), StateOK)
	}

	report, err := monitor.Observe(sampleWithSkew(sampleOrigin.Add(time.Hour), 45*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if report.State != StateOK {
		t.Fatalf("State = %s, want %s: a single outlier must not trip the alarm", report.State, StateOK)
	}
	if report.Skew != 45*time.Second {
		t.Fatalf("Skew = %s, want 45s", report.Skew)
	}
	if report.MaxSkew != 45*time.Second || report.MinSkew != 0 {
		t.Fatalf("range = (%s, %s)", report.MinSkew, report.MaxSkew)
	}
}

func TestMonitorEscalatesOnSustainedSkew(t *testing.T) {
	monitor := testMonitor()

	var last Report
	for i := 0; i < 6; i++ {
		report, err := monitor.Observe(sampleWithSkew(sampleOrigin.Add(time.Duration(i)*time.Minute), 40*time.Second))
		if err != nil {
			t.Fatal(err)
		}
		last = report
	}

	if last.State != StateCritical {
		t.Fatalf("State = %s, want %s", last.State, StateCritical)
	}
	if last.MedianSkew != 40*time.Second {
		t.Fatalf("MedianSkew = %s, want 40s", last.MedianSkew)
	}
	if monitor.Report().State != StateCritical {
		t.Fatalf("Report() = %+v", monitor.Report())
	}
}

func TestMonitorTransitionsAreReportedOnce(t *testing.T) {
	monitor := testMonitor()

	first, err := monitor.Observe(sampleWithSkew(sampleOrigin, 0))
	if err != nil {
		t.Fatal(err)
	}
	if !first.Transitioned {
		t.Fatal("first observation must report a transition from unknown")
	}

	second, err := monitor.Observe(sampleWithSkew(sampleOrigin.Add(time.Minute), 0))
	if err != nil {
		t.Fatal(err)
	}
	if second.Transitioned {
		t.Fatalf("steady state must not report a transition: %s -> %s", second.PreviousState, second.State)
	}
}

func TestMonitorReportsDrift(t *testing.T) {
	monitor := testMonitor()

	for i := 0; i < 4; i++ {
		at := sampleOrigin.Add(time.Duration(i) * time.Hour)
		if _, err := monitor.Observe(sampleWithSkew(at, time.Duration(i)*time.Second)); err != nil {
			t.Fatal(err)
		}
	}

	report := monitor.Report()
	if !report.Drift.Determined {
		t.Fatal("drift must be determined")
	}
	if delta := absDuration(report.Drift.PerDay - 24*time.Second); delta > time.Millisecond {
		t.Fatalf("Drift.PerDay = %s, want 24s", report.Drift.PerDay)
	}
}

func TestMonitorSetThresholdsNormalizes(t *testing.T) {
	monitor := testMonitor()

	applied := monitor.SetThresholds(Thresholds{Warn: 0, Critical: 0})
	if applied != DefaultThresholds() {
		t.Fatalf("SetThresholds() = %+v, want defaults", applied)
	}
	if monitor.Thresholds() != DefaultThresholds() {
		t.Fatalf("Thresholds() = %+v", monitor.Thresholds())
	}
}

func TestMonitorReset(t *testing.T) {
	monitor := testMonitor()
	if _, err := monitor.Observe(sampleWithSkew(sampleOrigin, 40*time.Second)); err != nil {
		t.Fatal(err)
	}

	monitor.Reset()
	if monitor.State() != StateUnknown {
		t.Fatalf("State() = %s, want %s", monitor.State(), StateUnknown)
	}
	if monitor.Report().Observed() {
		t.Fatalf("Report() = %+v, want empty", monitor.Report())
	}
	observed, _ := monitor.Counters()
	if observed != 1 {
		t.Fatalf("Counters() observed = %d, want 1: reset must not clear lifetime counters", observed)
	}
}

func TestMonitorWindowBoundsSampleCount(t *testing.T) {
	monitor := NewMonitor(DefaultThresholds(), 4)

	for i := 0; i < 10; i++ {
		if _, err := monitor.Observe(sampleWithSkew(sampleOrigin.Add(time.Duration(i)*time.Minute), 0)); err != nil {
			t.Fatal(err)
		}
	}
	if got := monitor.Report().SampleCount; got != 4 {
		t.Fatalf("SampleCount = %d, want 4", got)
	}
}

func TestMonitorConcurrentObservations(t *testing.T) {
	monitor := NewMonitor(DefaultThresholds(), 32)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				at := sampleOrigin.Add(time.Duration(index*25+j) * time.Second)
				if _, err := monitor.Observe(sampleWithSkew(at, time.Duration(j)*time.Millisecond)); err != nil {
					t.Errorf("Observe() error = %v", err)
					return
				}
				_ = monitor.Report()
				_ = monitor.State()
			}
		}(i)
	}
	wg.Wait()

	observed, rejected := monitor.Counters()
	if observed != 200 || rejected != 0 {
		t.Fatalf("Counters() = (%d, %d), want (200, 0)", observed, rejected)
	}
}
