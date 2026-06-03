package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
	gatewaygrpc "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/gateway"
	grpcserver "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/server"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/gateway"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/logging"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.LoadGateway()
	logger := logging.New(cfg.LogFile)
	logger.Printf("starting ft12 gateway (target=%s crc=%s interval=%s timeout=%s grpc=%s)",
		cfg.Target, cfg.CRCMode, cfg.PollInterval, cfg.RequestTimeout, cfg.GRPCListen)

	service, err := gateway.NewService(cfg, logger)
	if err != nil {
		logger.Fatalf("gateway config failed: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	errCh := make(chan error, 2)
	service.Start(ctx)

	if cfg.GRPCListen != "" {
		control := grpcserver.New(cfg.GRPCListen, func(s *grpc.Server) {
			ft12v1.RegisterGatewayServiceServer(s, gatewaygrpc.New(ctx, service))
		})
		logger.Printf("gateway gRPC control listening on %s", cfg.GRPCListen)
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
	if err := service.Stop(); err != nil {
		logger.Printf("gateway polling stopped with error: %v", err)
	}
	logger.Println("gateway stopped")
}
