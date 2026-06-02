package tcp

import (
	"context"
	"io"
	"net"
	"testing"
	"time"
)

func TestServerRunAcceptsConnectionAndStops(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := NewServer(ServerConfig{Address: "127.0.0.1:0"}, HandlerFunc(func(ctx context.Context, conn net.Conn) {
		buf := make([]byte, 4)
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		_, _ = conn.Write(buf[:n])
	}), nil)

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Run(ctx) }()

	addr := waitAddr(t, srv)
	conn, err := net.DialTimeout("tcp", addr.String(), time.Second)
	if err != nil {
		t.Fatalf("DialTimeout() error = %v", err)
	}
	defer conn.Close()

	if _, err := conn.Write([]byte("ping")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	out := make([]byte, 4)
	if _, err := io.ReadFull(conn, out); err != nil {
		t.Fatalf("ReadFull() error = %v", err)
	}
	if string(out) != "ping" {
		t.Fatalf("echo = %q, want ping", string(out))
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not stop")
	}
}

func waitAddr(t *testing.T, srv *Server) net.Addr {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if addr := srv.Addr(); addr != nil {
			return addr
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("server did not publish address")
	return nil
}
