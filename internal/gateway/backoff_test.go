package gateway

import (
	"testing"
	"time"
)

func TestBackoffNextAndReset(t *testing.T) {
	b := NewBackoff(100*time.Millisecond, 250*time.Millisecond)
	if got := b.Next(); got != 100*time.Millisecond {
		t.Fatalf("first = %s", got)
	}
	if got := b.Next(); got != 200*time.Millisecond {
		t.Fatalf("second = %s", got)
	}
	if got := b.Next(); got != 250*time.Millisecond {
		t.Fatalf("capped = %s", got)
	}
	b.Reset()
	if got := b.Next(); got != 100*time.Millisecond {
		t.Fatalf("after reset = %s", got)
	}
}

func TestBackoffNormalizesInvalidInput(t *testing.T) {
	b := NewBackoff(0, -1)
	if got := b.Next(); got != 500*time.Millisecond {
		t.Fatalf("default initial = %s", got)
	}

	b = NewBackoff(2*time.Second, time.Second)
	if got := b.Next(); got != 2*time.Second {
		t.Fatalf("max below initial should clamp to initial, got %s", got)
	}
	if got := b.Next(); got != 2*time.Second {
		t.Fatalf("capped = %s", got)
	}
}
