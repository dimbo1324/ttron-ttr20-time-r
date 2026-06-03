package gateway

import (
	"context"
	"io"
	"log"
	"net"
	"testing"
	"time"

	grpcclient "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/client"
	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
	grpcserver "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/server"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	emudomain "github.com/dimbo1324/ttron-ttr20-time-r/internal/emulator"
	gwdomain "github.com/dimbo1324/ttron-ttr20-time-r/internal/gateway"
	"google.golang.org/grpc"
)

func TestGatewayGRPCService(t *testing.T) {
	emu := emudomain.NewServer(&config.EmulatorConfig{
		Listen:               "127.0.0.1:0",
		CRCMode:              "sum",
		AdapterAddr:          1,
		ReadTimeoutDuration:  time.Second,
		WriteTimeoutDuration: time.Second,
		RecentSize:           10,
	}, log.New(io.Discard, "", 0))
	emuErr := make(chan error, 1)
	go func() { emuErr <- emu.Start() }()
	waitFor(t, time.Second, func() bool { return emu.Addr() != nil })
	defer stopEmulator(t, emu, emuErr)

	service, err := gwdomain.NewService(&config.GatewayConfig{
		Target:         emu.Addr().String(),
		CRCMode:        "sum",
		AdapterAddr:    1,
		PollInterval:   50 * time.Millisecond,
		RequestTimeout: 500 * time.Millisecond,
		ConnectTimeout: 500 * time.Millisecond,
		BackoffInitial: 10 * time.Millisecond,
		BackoffMax:     50 * time.Millisecond,
		RecentSize:     10,
	}, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatal(err)
	}

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	service.Start(rootCtx)

	server := grpcserver.New("127.0.0.1:0", func(s *grpc.Server) {
		ft12v1.RegisterGatewayServiceServer(s, New(rootCtx, service))
	})
	serverErr := make(chan error, 1)
	go func() { serverErr <- server.Run(rootCtx) }()
	addr := waitGRPCAddr(t, server)

	client, conn, err := grpcclient.DialGateway(rootCtx, addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	waitFor(t, 2*time.Second, func() bool {
		callCtx, callCancel := context.WithTimeout(rootCtx, time.Second)
		defer callCancel()
		status, err := client.GetStatus(callCtx, &ft12v1.GetGatewayStatusRequest{})
		return err == nil && status.GetStatus().GetSuccessfulReads() > 0
	})

	callCtx, callCancel := context.WithTimeout(rootCtx, time.Second)
	defer callCancel()
	last, err := client.GetLastReadTime(callCtx, &ft12v1.GetLastReadTimeRequest{})
	if err != nil {
		t.Fatalf("GetLastReadTime() error = %v", err)
	}
	if !last.GetAvailable() || last.GetDeviceTime() == nil {
		t.Fatalf("last read = %+v", last)
	}

	if _, err := client.StopPolling(callCtx, &ft12v1.StopPollingRequest{}); err != nil {
		t.Fatalf("StopPolling() error = %v", err)
	}
	stopped, err := client.GetStatus(callCtx, &ft12v1.GetGatewayStatusRequest{})
	if err != nil {
		t.Fatalf("GetStatus() after stop error = %v", err)
	}
	if stopped.GetStatus().GetState() != ft12v1.ServiceState_SERVICE_STATE_STOPPED {
		t.Fatalf("status after stop = %+v", stopped.GetStatus())
	}
	if _, err := client.StartPolling(callCtx, &ft12v1.StartPollingRequest{}); err != nil {
		t.Fatalf("StartPolling() error = %v", err)
	}
	if _, err := client.GetRecentEvents(callCtx, &ft12v1.GetRecentEventsRequest{Limit: 5}); err != nil {
		t.Fatalf("GetRecentEvents() error = %v", err)
	}

	cancel()
	_ = service.Stop()
	select {
	case err := <-serverErr:
		if err != nil {
			t.Fatalf("gRPC server stopped with error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("gRPC server did not stop")
	}
}

func waitGRPCAddr(t *testing.T, provider interface{ Addr() net.Addr }) string {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if addr := provider.Addr(); addr != nil {
			return addr.String()
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("gRPC server did not publish address")
	return ""
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

func stopEmulator(t *testing.T, srv *emudomain.Server, errCh chan error) {
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
