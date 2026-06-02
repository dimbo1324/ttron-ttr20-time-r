package gateway

import (
	"context"
	"io"
	"log"
	"testing"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/emulator"
)

func TestGatewayPollsEmulator(t *testing.T) {
	emu, emuErrCh := startGatewayTestEmulator(t, "sum")
	defer stopGatewayTestEmulator(t, emu, emuErrCh)

	cfg := &config.GatewayConfig{
		Target:         emu.Addr().String(),
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
	emuCfg := &config.EmulatorConfig{Listen: "127.0.0.1:0", CRCMode: "sum", AdapterAddr: 1, NoResponse: true, ReadTimeoutDuration: time.Second, WriteTimeoutDuration: time.Second, RecentSize: 10}
	emu := emulator.NewServer(emuCfg, log.New(io.Discard, "", 0))
	emuErrCh := make(chan error, 1)
	go func() { emuErrCh <- emu.Start() }()
	waitFor(t, time.Second, func() bool { return emu.Addr() != nil })
	defer stopGatewayTestEmulator(t, emu, emuErrCh)

	cfg := &config.GatewayConfig{
		Target:         emu.Addr().String(),
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

func startGatewayTestEmulator(t *testing.T, mode string) (*emulator.Server, chan error) {
	t.Helper()
	cfg := &config.EmulatorConfig{Listen: "127.0.0.1:0", CRCMode: mode, AdapterAddr: 1, ReadTimeoutDuration: time.Second, WriteTimeoutDuration: time.Second, RecentSize: 10}
	srv := emulator.NewServer(cfg, log.New(io.Discard, "", 0))
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Start() }()
	waitFor(t, time.Second, func() bool { return srv.Addr() != nil })
	return srv, errCh
}

func stopGatewayTestEmulator(t *testing.T, srv *emulator.Server, errCh chan error) {
	t.Helper()
	srv.Stop()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("emulator stopped with error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("emulator did not stop")
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
