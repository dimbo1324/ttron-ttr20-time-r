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
