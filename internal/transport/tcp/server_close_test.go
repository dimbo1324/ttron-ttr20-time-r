package tcp

import (
	"context"
	"io"
	"log"
	"net"
	"testing"
	"time"
)

type noopHandler struct{}

func (noopHandler) HandleConnection(context.Context, net.Conn) {}

func TestServerCloseBeforeRun(t *testing.T) {
	server := NewServer(ServerConfig{Address: "127.0.0.1:0"}, noopHandler{}, log.New(io.Discard, "", 0))

	if err := server.Close(); err != nil {
		t.Fatalf("Close() before Run() = %v", err)
	}
	if addr := server.Addr(); addr != nil {
		t.Fatalf("Addr() = %v, want nil before Run()", addr)
	}
}

func TestServerCloseStopsRun(t *testing.T) {
	server := NewServer(ServerConfig{Address: "127.0.0.1:0"}, noopHandler{}, log.New(io.Discard, "", 0))

	errCh := make(chan error, 1)
	go func() { errCh <- server.Run(context.Background()) }()

	deadline := time.Now().Add(2 * time.Second)
	for server.Addr() == nil {
		if time.Now().After(deadline) {
			t.Fatal("server did not start listening")
		}
		time.Sleep(10 * time.Millisecond)
	}

	if err := server.Close(); err != nil {
		t.Fatalf("Close() = %v", err)
	}
	select {
	case <-errCh:
	case <-time.After(3 * time.Second):
		t.Fatal("Run() did not return after Close()")
	}
}

func TestDialRejectsUnreachableTarget(t *testing.T) {
	_, err := Dial(context.Background(), ClientConfig{Address: "127.0.0.1:1", ConnectTimeout: 100 * time.Millisecond})
	if err == nil {
		t.Fatal("Dial() must fail for an unreachable target")
	}
}

func TestDialRejectsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := Dial(ctx, ClientConfig{Address: "127.0.0.1:1", ConnectTimeout: time.Second}); err == nil {
		t.Fatal("Dial() must fail for a cancelled context")
	}
}
