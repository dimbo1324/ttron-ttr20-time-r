package events

import (
	"sync"
	"testing"
	"time"
)

func TestNewRingCapacityFallback(t *testing.T) {
	for _, capacity := range []int{0, -5} {
		ring := NewRing(capacity)
		for i := 0; i < 101; i++ {
			ring.Add(FrameRecord{Direction: DirectionTX})
		}
		if got := ring.Len(); got != 100 {
			t.Fatalf("NewRing(%d).Len() = %d, want the default capacity 100", capacity, got)
		}
	}
}

func TestRingLenTracksGrowthAndSaturation(t *testing.T) {
	ring := NewRing(3)
	if got := ring.Len(); got != 0 {
		t.Fatalf("Len() = %d, want 0", got)
	}

	for i := 1; i <= 3; i++ {
		ring.Add(FrameRecord{Direction: DirectionRX})
		if got := ring.Len(); got != i {
			t.Fatalf("Len() = %d, want %d", got, i)
		}
	}

	ring.Add(FrameRecord{Direction: DirectionRX})
	if got := ring.Len(); got != 3 {
		t.Fatalf("Len() = %d, want the capacity 3", got)
	}
}

func TestRingAssignsIncreasingIdentifiers(t *testing.T) {
	ring := NewRing(4)
	for i := 0; i < 3; i++ {
		ring.Add(FrameRecord{Direction: DirectionTX})
	}

	snapshot := ring.Snapshot()
	for i, record := range snapshot {
		if record.ID != uint64(i+1) {
			t.Fatalf("Snapshot()[%d].ID = %d, want %d", i, record.ID, i+1)
		}
		if record.Timestamp.IsZero() {
			t.Fatal("Add() must stamp records that arrive without a timestamp")
		}
	}
}

func TestRingKeepsExplicitTimestamp(t *testing.T) {
	moment := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	ring := NewRing(2)
	ring.Add(FrameRecord{Timestamp: moment, Direction: DirectionSystem})

	if got := ring.Snapshot()[0].Timestamp; !got.Equal(moment) {
		t.Fatalf("Timestamp = %s, want %s", got, moment)
	}
}

func TestRingConcurrentAccess(t *testing.T) {
	ring := NewRing(64)
	var wg sync.WaitGroup

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				ring.Add(FrameRecord{Direction: DirectionTX})
				_ = ring.Snapshot()
				_ = ring.Len()
			}
		}()
	}
	wg.Wait()

	if got := ring.Len(); got != 64 {
		t.Fatalf("Len() = %d, want 64", got)
	}
}
