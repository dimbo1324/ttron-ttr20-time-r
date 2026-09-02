package health

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func testPolicy() Policy {
	return Policy{DegradeAfter: 3, OfflineAfter: 5, RecoverAfter: 2, WindowSize: 16}
}

func TestTrackerStartsUnknown(t *testing.T) {
	tracker := NewTracker(testPolicy())

	if tracker.State() != StateUnknown {
		t.Fatalf("State() = %s, want %s", tracker.State(), StateUnknown)
	}
	snapshot := tracker.Snapshot()
	if snapshot.TotalSuccesses != 0 || snapshot.TotalFailures != 0 {
		t.Fatalf("Snapshot() = %+v", snapshot)
	}
}

func TestTrackerFirstSuccessGoesOnline(t *testing.T) {
	tracker := NewTracker(testPolicy())

	transition := tracker.RecordSuccess(15 * time.Millisecond)
	if !transition.Changed || transition.From != StateUnknown || transition.To != StateOnline {
		t.Fatalf("transition = %+v", transition)
	}
	if tracker.State() != StateOnline {
		t.Fatalf("State() = %s", tracker.State())
	}

	snapshot := tracker.Snapshot()
	if snapshot.TotalSuccesses != 1 || snapshot.ConsecutiveSuccesses != 1 {
		t.Fatalf("Snapshot() = %+v", snapshot)
	}
	if snapshot.Availability != 1 {
		t.Fatalf("Availability = %f, want 1", snapshot.Availability)
	}
	if snapshot.LastSuccessAt.IsZero() || !snapshot.Since.Equal(snapshot.LastSuccessAt) {
		t.Fatalf("Since = %s, LastSuccessAt = %s", snapshot.Since, snapshot.LastSuccessAt)
	}
}

func TestTrackerHysteresisOnFailures(t *testing.T) {
	tracker := NewTracker(testPolicy())
	tracker.RecordSuccess(time.Millisecond)

	failure := errors.New("boom")
	for i := 1; i <= 2; i++ {
		transition := tracker.RecordFailure(failure)
		if transition.Changed {
			t.Fatalf("failure %d must not change the state yet: %+v", i, transition)
		}
		if tracker.State() != StateOnline {
			t.Fatalf("after %d failures State() = %s, want %s", i, tracker.State(), StateOnline)
		}
	}

	transition := tracker.RecordFailure(failure)
	if !transition.Changed || transition.To != StateDegraded {
		t.Fatalf("third failure transition = %+v, want degraded", transition)
	}

	for i := 4; i <= 4; i++ {
		if transition := tracker.RecordFailure(failure); transition.Changed {
			t.Fatalf("failure %d must keep the degraded state: %+v", i, transition)
		}
	}

	transition = tracker.RecordFailure(failure)
	if !transition.Changed || transition.To != StateOffline {
		t.Fatalf("fifth failure transition = %+v, want offline", transition)
	}
	if transition.Reason != failure.Error() {
		t.Fatalf("Reason = %q, want %q", transition.Reason, failure.Error())
	}

	snapshot := tracker.Snapshot()
	if snapshot.ConsecutiveFailures != 5 || snapshot.LastError != failure.Error() {
		t.Fatalf("Snapshot() = %+v", snapshot)
	}
}

func TestTrackerRecoveryRequiresConsecutiveSuccesses(t *testing.T) {
	tracker := NewTracker(testPolicy())
	failure := errors.New("down")
	for i := 0; i < 5; i++ {
		tracker.RecordFailure(failure)
	}
	if tracker.State() != StateOffline {
		t.Fatalf("State() = %s, want %s", tracker.State(), StateOffline)
	}

	if transition := tracker.RecordSuccess(time.Millisecond); transition.Changed {
		t.Fatalf("a single success must not recover the device: %+v", transition)
	}
	if tracker.State() != StateOffline {
		t.Fatalf("State() = %s, want %s", tracker.State(), StateOffline)
	}

	transition := tracker.RecordSuccess(time.Millisecond)
	if !transition.Changed || transition.To != StateOnline {
		t.Fatalf("transition = %+v, want online", transition)
	}
	if tracker.Snapshot().LastError != "" {
		t.Fatalf("LastError must be cleared after recovery")
	}
}

func TestTrackerFailureStreakResetsOnSuccess(t *testing.T) {
	tracker := NewTracker(testPolicy())
	tracker.RecordSuccess(time.Millisecond)

	failure := errors.New("flap")
	tracker.RecordFailure(failure)
	tracker.RecordFailure(failure)
	tracker.RecordSuccess(time.Millisecond)

	if got := tracker.Snapshot().ConsecutiveFailures; got != 0 {
		t.Fatalf("ConsecutiveFailures = %d, want 0", got)
	}
	for i := 0; i < 2; i++ {
		tracker.RecordFailure(failure)
	}
	if tracker.State() != StateOnline {
		t.Fatalf("State() = %s, want %s: the streak must have restarted", tracker.State(), StateOnline)
	}
}

func TestTrackerStaysUnknownWhileNeverReached(t *testing.T) {
	tracker := NewTracker(testPolicy())
	failure := errors.New("unreachable")

	for i := 0; i < 2; i++ {
		if transition := tracker.RecordFailure(failure); transition.Changed {
			t.Fatalf("transition = %+v, want no change", transition)
		}
	}
	if tracker.State() != StateUnknown {
		t.Fatalf("State() = %s, want %s", tracker.State(), StateUnknown)
	}

	tracker.RecordFailure(failure)
	if tracker.State() != StateDegraded {
		t.Fatalf("State() = %s, want %s", tracker.State(), StateDegraded)
	}
}

func TestTrackerSnapshotReportsLatency(t *testing.T) {
	tracker := NewTracker(testPolicy())
	for _, latency := range []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond} {
		tracker.RecordSuccess(latency)
	}

	snapshot := tracker.Snapshot()
	if snapshot.Latency.Samples != 3 {
		t.Fatalf("Latency.Samples = %d, want 3", snapshot.Latency.Samples)
	}
	if snapshot.Latency.P50 != 20*time.Millisecond {
		t.Fatalf("Latency.P50 = %s, want 20ms", snapshot.Latency.P50)
	}
	if snapshot.Policy != testPolicy() {
		t.Fatalf("Snapshot().Policy = %+v", snapshot.Policy)
	}
}

func TestTrackerWithClock(t *testing.T) {
	moment := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	tracker := NewTracker(testPolicy()).WithClock(func() time.Time { return moment })

	tracker.RecordSuccess(time.Millisecond)
	if got := tracker.Snapshot().LastSuccessAt; !got.Equal(moment) {
		t.Fatalf("LastSuccessAt = %s, want %s", got, moment)
	}

	tracker.WithClock(nil)
	if got := tracker.Snapshot().LastSuccessAt; !got.Equal(moment) {
		t.Fatalf("WithClock(nil) must keep the previous clock")
	}
}

func TestTrackerRecordFailureWithNilError(t *testing.T) {
	tracker := NewTracker(testPolicy())
	transition := tracker.RecordFailure(nil)

	if transition.Reason != "failed read" {
		t.Fatalf("Reason = %q", transition.Reason)
	}
	if got := tracker.Snapshot().LastError; got != "" {
		t.Fatalf("LastError = %q, want empty", got)
	}
}

func TestTrackerReset(t *testing.T) {
	tracker := NewTracker(testPolicy())
	tracker.RecordSuccess(time.Millisecond)
	tracker.RecordFailure(errors.New("boom"))

	tracker.Reset()
	snapshot := tracker.Snapshot()
	if snapshot.State != StateUnknown || snapshot.ConsecutiveFailures != 0 || snapshot.LastError != "" {
		t.Fatalf("Snapshot() after reset = %+v", snapshot)
	}
	if snapshot.WindowSamples != 0 {
		t.Fatalf("WindowSamples = %d, want 0", snapshot.WindowSamples)
	}
	if snapshot.TotalSuccesses != 1 || snapshot.TotalFailures != 1 {
		t.Fatalf("reset must keep lifetime counters: %+v", snapshot)
	}
}

func TestTrackerConcurrentRecording(t *testing.T) {
	tracker := NewTracker(Policy{DegradeAfter: 3, OfflineAfter: 5, RecoverAfter: 2, WindowSize: 64})

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				if (index+j)%2 == 0 {
					tracker.RecordSuccess(time.Duration(j) * time.Millisecond)
				} else {
					tracker.RecordFailure(errors.New("noise"))
				}
				_ = tracker.Snapshot()
				_ = tracker.State()
			}
		}(i)
	}
	wg.Wait()

	snapshot := tracker.Snapshot()
	if total := snapshot.TotalSuccesses + snapshot.TotalFailures; total != 200 {
		t.Fatalf("total outcomes = %d, want 200", total)
	}
}
