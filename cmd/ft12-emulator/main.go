package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	emulatorgrpc "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/emulator"
	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
	grpcserver "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/server"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/emulator"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/logging"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.LoadEmulator()
	logger := logging.New(cfg.LogFile)
	logger.Printf("starting ft12 emulator (listen=%s crc=%s adapter=%d grpc=%s)",
		cfg.ListenAddress(), cfg.CRCMode, cfg.AdapterAddr, cfg.GRPCListen)

	service, err := emulator.NewService(cfg, logger)
	if err != nil {
		logger.Fatalf("emulator config failed: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	errCh := make(chan error, 2)
	go func() {
		errCh <- service.Run(ctx)
	}()

	if cfg.GRPCListen != "" {
		control := grpcserver.New(cfg.GRPCListen, func(s *grpc.Server) {
			ft12v1.RegisterEmulatorServiceServer(s, emulatorgrpc.New(service))
		})
		logger.Printf("emulator gRPC control listening on %s", cfg.GRPCListen)
		go func() {
			errCh <- control.Run(ctx)
		}()
	}

	select {
	case <-ctx.Done():
		logger.Printf("signal received, shutting down...")
	case err := <-errCh:
		if err != nil {
			logger.Printf("service stopped with error: %v", err)
		}
		cancel()
	}
	logger.Println("bye")
}
