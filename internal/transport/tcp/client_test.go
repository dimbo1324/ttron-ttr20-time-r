package tcp

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestDial(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	accepted := make(chan struct{})
	go func() {
		conn, err := ln.Accept()
		if err == nil {
			conn.Close()
		}
		close(accepted)
	}()

	conn, err := Dial(context.Background(), ClientConfig{Address: ln.Addr().String(), ConnectTimeout: time.Second})
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	conn.Close()

	select {
	case <-accepted:
	case <-time.After(time.Second):
		t.Fatal("server did not accept connection")
	}
}
