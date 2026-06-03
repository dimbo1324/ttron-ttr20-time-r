package emulator

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
	domain "github.com/dimbo1324/ttron-ttr20-time-r/internal/emulator"
	"google.golang.org/grpc"
)

func TestEmulatorGRPCService(t *testing.T) {
	service, err := domain.NewService(&config.EmulatorConfig{
		Listen:               "127.0.0.1:0",
		CRCMode:              "sum",
		AdapterAddr:          1,
		ReadTimeoutDuration:  time.Second,
		WriteTimeoutDuration: time.Second,
		RecentSize:           10,
	}, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server := grpcserver.New("127.0.0.1:0", func(s *grpc.Server) {
		ft12v1.RegisterEmulatorServiceServer(s, New(service))
	})
	errCh := make(chan error, 1)
	go func() { errCh <- server.Run(ctx) }()
	addr := waitGRPCAddr(t, server)

	client, conn, err := grpcclient.DialEmulator(ctx, addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	callCtx, callCancel := context.WithTimeout(ctx, time.Second)
	defer callCancel()
	status, err := client.GetStatus(callCtx, &ft12v1.GetEmulatorStatusRequest{})
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if status.GetStatus().GetChecksumMode() != ft12v1.ChecksumMode_CHECKSUM_MODE_SUM {
		t.Fatalf("status = %+v", status.GetStatus())
	}

	set, err := client.SetFaultMode(callCtx, &ft12v1.SetFaultModeRequest{FaultMode: &ft12v1.FaultMode{
		CorruptChecksum:   true,
		FragmentResponse:  true,
		FragmentDelayMs:   5,
		ResponseDelayMs:   10,
		NoResponse:        true,
		CloseAfterRequest: true,
	}})
	if err != nil {
		t.Fatalf("SetFaultMode() error = %v", err)
	}
	if !set.GetFaultMode().GetCorruptChecksum() || !set.GetFaultMode().GetNoResponse() {
		t.Fatalf("fault = %+v", set.GetFaultMode())
	}

	fault, err := client.GetFaultMode(callCtx, &ft12v1.GetFaultModeRequest{})
	if err != nil {
		t.Fatalf("GetFaultMode() error = %v", err)
	}
	if !fault.GetFaultMode().GetFragmentResponse() || fault.GetFaultMode().GetFragmentDelayMs() != 5 {
		t.Fatalf("fault = %+v", fault.GetFaultMode())
	}

	if _, err := client.GetRecentEvents(callCtx, &ft12v1.GetRecentEventsRequest{Limit: 10}); err != nil {
		t.Fatalf("GetRecentEvents() error = %v", err)
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("server stopped with error: %v", err)
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
