package clock

import (
	"math"
	"testing"
	"time"
)

func hourlyEntries(step time.Duration, skews ...time.Duration) []entry {
	out := make([]entry, 0, len(skews))
	for i, skew := range skews {
		out = append(out, entry{at: sampleOrigin.Add(time.Duration(i) * step), skew: skew})
	}
	return out
}

func TestComputeDriftRequiresMinimumSamples(t *testing.T) {
	for count := 0; count < MinDriftSamples; count++ {
		skews := make([]time.Duration, count)
		drift := computeDrift(hourlyEntries(time.Hour, skews...))
		if drift.Determined {
			t.Fatalf("drift determined with %d samples", count)
		}
		if drift.SampleCount != count {
			t.Fatalf("SampleCount = %d, want %d", drift.SampleCount, count)
		}
	}
}

func TestComputeDriftLinearGain(t *testing.T) {
	items := hourlyEntries(time.Hour, 0, time.Second, 2*time.Second, 3*time.Second)
	drift := computeDrift(items)

	if !drift.Determined {
		t.Fatal("drift must be determined for a linear series")
	}
	if drift.SampleCount != 4 {
		t.Fatalf("SampleCount = %d, want 4", drift.SampleCount)
	}
	if drift.Span != 3*time.Hour {
		t.Fatalf("Span = %s, want 3h", drift.Span)
	}
	want := 24 * time.Second
	if delta := absDuration(drift.PerDay - want); delta > time.Millisecond {
		t.Fatalf("PerDay = %s, want %s", drift.PerDay, want)
	}
	if math.Abs(drift.Fit-1) > 1e-9 {
		t.Fatalf("Fit = %f, want 1", drift.Fit)
	}
}

func TestComputeDriftNegativeSlope(t *testing.T) {
	items := hourlyEntries(30*time.Minute, 0, -time.Second, -2*time.Second, -3*time.Second)
	drift := computeDrift(items)

	want := -48 * time.Second
	if delta := absDuration(drift.PerDay - want); delta > time.Millisecond {
		t.Fatalf("PerDay = %s, want %s", drift.PerDay, want)
	}
}

func TestComputeDriftStableClock(t *testing.T) {
	items := hourlyEntries(time.Hour, time.Second, time.Second, time.Second, time.Second)
	drift := computeDrift(items)

	if !drift.Determined {
		t.Fatal("drift must be determined for a flat series")
	}
	if drift.PerDay != 0 {
		t.Fatalf("PerDay = %s, want 0", drift.PerDay)
	}
	if drift.Fit != 0 {
		t.Fatalf("Fit = %f, want 0 for a zero-variance series", drift.Fit)
	}
}

func TestComputeDriftWithoutTimeSpread(t *testing.T) {
	items := []entry{
		{at: sampleOrigin, skew: time.Second},
		{at: sampleOrigin, skew: 2 * time.Second},
		{at: sampleOrigin, skew: 3 * time.Second},
	}
	drift := computeDrift(items)

	if drift.Determined {
		t.Fatal("drift must not be determined without time spread")
	}
	if drift.SampleCount != 3 {
		t.Fatalf("SampleCount = %d, want 3", drift.SampleCount)
	}
}

func TestComputeDriftClampsExtremeSlope(t *testing.T) {
	items := []entry{
		{at: sampleOrigin, skew: 0},
		{at: sampleOrigin.Add(time.Nanosecond), skew: time.Hour},
		{at: sampleOrigin.Add(2 * time.Nanosecond), skew: 2 * time.Hour},
	}
	drift := computeDrift(items)

	if !drift.Determined {
		t.Fatal("drift must be determined")
	}
	if absDuration(drift.PerDay) > time.Duration(maxDriftPerDayAbs) {
		t.Fatalf("PerDay = %s exceeds the clamp", drift.PerDay)
	}
}

func TestComputeDriftNoisySeriesLowersFit(t *testing.T) {
	items := hourlyEntries(time.Hour, 0, 5*time.Second, -5*time.Second, 10*time.Second, -10*time.Second)
	drift := computeDrift(items)

	if !drift.Determined {
		t.Fatal("drift must be determined")
	}
	if drift.Fit > 0.5 {
		t.Fatalf("Fit = %f, want a low value for a noisy series", drift.Fit)
	}
}
