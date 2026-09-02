package emulator

import (
	"sync"
	"testing"
	"time"
)

type stubClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *stubClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *stubClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func TestDeviceClockWithoutFaults(t *testing.T) {
	base := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	source := &stubClock{now: base}
	deviceClock := NewDeviceClock(source.Now, 0, 0)

	if got := deviceClock.Now(); !got.Equal(base) {
		t.Fatalf("Now() = %s, want %s", got, base)
	}

	source.advance(time.Hour)
	if got := deviceClock.Now(); !got.Equal(base.Add(time.Hour)) {
		t.Fatalf("Now() = %s, want %s", got, base.Add(time.Hour))
	}
}

func TestDeviceClockAppliesConstantOffset(t *testing.T) {
	base := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	source := &stubClock{now: base}

	tests := []struct {
		name   string
		offset time.Duration
	}{
		{name: "ahead", offset: 45 * time.Second},
		{name: "behind", offset: -90 * time.Second},
		{name: "none", offset: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deviceClock := NewDeviceClock(source.Now, tt.offset, 0)
			if got := deviceClock.Now(); !got.Equal(base.Add(tt.offset)) {
				t.Fatalf("Now() = %s, want %s", got, base.Add(tt.offset))
			}
		})
	}
}

func TestDeviceClockAccumulatesDrift(t *testing.T) {
	base := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	source := &stubClock{now: base}
	deviceClock := NewDeviceClock(source.Now, 0, 24*time.Second)

	source.advance(12 * time.Hour)
	got := deviceClock.Now()
	want := base.Add(12*time.Hour + 12*time.Second)
	if delta := got.Sub(want); delta > time.Millisecond || delta < -time.Millisecond {
		t.Fatalf("Now() = %s, want about %s", got, want)
	}

	source.advance(12 * time.Hour)
	got = deviceClock.Now()
	want = base.Add(24*time.Hour + 24*time.Second)
	if delta := got.Sub(want); delta > time.Millisecond || delta < -time.Millisecond {
		t.Fatalf("Now() = %s, want about %s", got, want)
	}
}

func TestDeviceClockCombinesOffsetAndDrift(t *testing.T) {
	base := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	source := &stubClock{now: base}
	deviceClock := NewDeviceClock(source.Now, time.Minute, -48*time.Second)

	source.advance(12 * time.Hour)
	got := deviceClock.Now()
	want := base.Add(12*time.Hour + time.Minute - 24*time.Second)
	if delta := got.Sub(want); delta > time.Millisecond || delta < -time.Millisecond {
		t.Fatalf("Now() = %s, want about %s", got, want)
	}
}

func TestDeviceClockIgnoresBackwardsSource(t *testing.T) {
	base := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	source := &stubClock{now: base}
	deviceClock := NewDeviceClock(source.Now, 0, 24*time.Hour)

	source.advance(-time.Hour)
	got := deviceClock.Now()
	if !got.Equal(base.Add(-time.Hour)) {
		t.Fatalf("Now() = %s, want no accumulated drift for a backwards source", got)
	}
}

func TestDeviceClockConfigureResetsDriftOrigin(t *testing.T) {
	base := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	source := &stubClock{now: base}
	deviceClock := NewDeviceClock(source.Now, 0, 24*time.Second)

	source.advance(24 * time.Hour)
	deviceClock.Configure(0, 48*time.Second)

	offset, drift := deviceClock.Settings()
	if offset != 0 || drift != 48*time.Second {
		t.Fatalf("Settings() = (%s, %s)", offset, drift)
	}

	if got := deviceClock.Now(); !got.Equal(source.Now()) {
		t.Fatalf("Now() = %s, want the drift clock restarted at %s", got, source.Now())
	}

	source.advance(24 * time.Hour)
	got := deviceClock.Now()
	want := source.Now().Add(48 * time.Second)
	if delta := got.Sub(want); delta > time.Millisecond || delta < -time.Millisecond {
		t.Fatalf("Now() = %s, want about %s", got, want)
	}
}

func TestDeviceClockConfigureKeepsOriginForSameDrift(t *testing.T) {
	base := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	source := &stubClock{now: base}
	deviceClock := NewDeviceClock(source.Now, 0, 24*time.Second)

	source.advance(24 * time.Hour)
	deviceClock.Configure(time.Minute, 24*time.Second)

	got := deviceClock.Now()
	want := source.Now().Add(time.Minute + 24*time.Second)
	if delta := got.Sub(want); delta > time.Millisecond || delta < -time.Millisecond {
		t.Fatalf("Now() = %s, want about %s", got, want)
	}
}

func TestNewDeviceClockDefaultsToWallClock(t *testing.T) {
	deviceClock := NewDeviceClock(nil, 0, 0)

	got := deviceClock.Now()
	if delta := time.Since(got); delta > time.Second || delta < -time.Second {
		t.Fatalf("Now() = %s, want the current time", got)
	}
}

func TestDeviceClockConcurrentUse(t *testing.T) {
	deviceClock := NewDeviceClock(time.Now, 0, 0)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				deviceClock.Configure(time.Duration(index)*time.Second, time.Duration(j)*time.Second)
				_ = deviceClock.Now()
				_, _ = deviceClock.Settings()
			}
		}(i)
	}
	wg.Wait()
}
