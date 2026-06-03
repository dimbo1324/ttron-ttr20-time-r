package lifecycle

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestGroupContextCancellationStopsRunners(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	group := NewGroup(nil)
	stopped := make(chan struct{})
	group.Add("waiter", func(ctx context.Context) error {
		<-ctx.Done()
		close(stopped)
		return nil
	})

	done := make(chan error, 1)
	go func() { done <- group.Run(ctx) }()
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("group did not stop after cancellation")
	}
	select {
	case <-stopped:
	default:
		t.Fatal("runner did not observe cancellation")
	}
}

func TestGroupRunnerErrorCancelsGroup(t *testing.T) {
	want := errors.New("boom")
	group := NewGroup(nil)
	cancelled := make(chan struct{})
	group.Add("failing", func(context.Context) error {
		return want
	})
	group.Add("waiter", func(ctx context.Context) error {
		<-ctx.Done()
		close(cancelled)
		return nil
	})

	err := group.Run(context.Background())
	if !errors.Is(err, want) {
		t.Fatalf("Run error = %v, want %v", err, want)
	}
	select {
	case <-cancelled:
	default:
		t.Fatal("sibling runner was not cancelled")
	}
}
