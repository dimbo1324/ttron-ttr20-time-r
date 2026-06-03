package gateway

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"testing"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/codec"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
)

func TestGatewayPollsEmulator(t *testing.T) {
	addr, stopDevice := startGatewayTestDevice(t, "sum", false)
	defer stopDevice()

	cfg := &config.GatewayConfig{
		Target:         addr,
		CRCMode:        "sum",
		AdapterAddr:    1,
		PollInterval:   50 * time.Millisecond,
		RequestTimeout: 500 * time.Millisecond,
		ConnectTimeout: 500 * time.Millisecond,
		BackoffInitial: 10 * time.Millisecond,
		BackoffMax:     50 * time.Millisecond,
		RecentSize:     10,
	}
	service, err := NewService(cfg, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- service.Run(ctx) }()

	waitFor(t, time.Second, func() bool {
		return service.Status().SuccessfulReads > 0
	})
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("gateway stopped with error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("gateway did not stop")
	}

	status := service.Status()
	if status.LastParsedDeviceTime.IsZero() || status.RecentFramesCount == 0 {
		t.Fatalf("status = %+v", status)
	}
}

func TestGatewayRecordsTimeoutFailure(t *testing.T) {
	addr, stopDevice := startGatewayTestDevice(t, "sum", true)
	defer stopDevice()

	cfg := &config.GatewayConfig{
		Target:         addr,
		CRCMode:        "sum",
		AdapterAddr:    1,
		PollInterval:   50 * time.Millisecond,
		RequestTimeout: 80 * time.Millisecond,
		ConnectTimeout: 500 * time.Millisecond,
		BackoffInitial: 10 * time.Millisecond,
		BackoffMax:     20 * time.Millisecond,
		RecentSize:     10,
	}
	service, err := NewService(cfg, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- service.Run(ctx) }()
	waitFor(t, time.Second, func() bool { return service.Status().FailedReads > 0 })
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("gateway stopped with error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("gateway did not stop")
	}
	if service.Status().LastError == "" {
		t.Fatalf("expected last error, status = %+v", service.Status())
	}
}

func startGatewayTestDevice(t *testing.T, modeName string, noResponse bool) (string, func()) {
	t.Helper()
	mode, err := checksum.ParseMode(modeName)
	if err != nil {
		t.Fatal(err)
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			conn, err := ln.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				}
				return
			}
			go handleGatewayTestDeviceConn(conn, mode, noResponse)
		}
	}()

	stop := func() {
		_ = ln.Close()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("test device did not stop")
		}
	}
	return ln.Addr().String(), stop
}

func handleGatewayTestDeviceConn(conn net.Conn, mode checksum.Mode, noResponse bool) {
	defer conn.Close()
	wire := codec.New(mode, 0x00, 0x01)
	parser := frame.NewStreamParser(mode)
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		result := parser.Push(buf[:n])
		if len(result.Frames) == 0 {
			continue
		}
		if noResponse {
			continue
		}
		resp, err := wire.EncodeReadTimeResponse(result.Frames[0], time.Now())
		if err != nil {
			return
		}
		if _, err := conn.Write(resp); err != nil {
			return
		}
	}
}

func waitFor(t *testing.T, timeout time.Duration, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition was not met")
}
