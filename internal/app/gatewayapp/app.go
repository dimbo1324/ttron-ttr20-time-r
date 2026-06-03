package gatewayapp

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
	gatewaygrpc "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/gateway"
	grpcserver "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/server"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/gateway"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/platform/lifecycle"
	platformlogging "github.com/dimbo1324/ttron-ttr20-time-r/internal/platform/logging"
	"google.golang.org/grpc"
)

func Run(args []string) int {
	cfg, err := config.LoadGateway(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "gateway config failed: %v\n", err)
		return 1
	}

	logger := platformlogging.New(cfg.LogFile)
	logger.Printf("starting ft12 gateway (target=%s crc=%s interval=%s timeout=%s grpc=%s)",
		cfg.Target, cfg.CRCMode, cfg.PollInterval, cfg.RequestTimeout, cfg.GRPCListen)

	service, err := gateway.NewService(cfg, logger)
	if err != nil {
		logger.Printf("gateway service creation failed: %v", err)
		return 1
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	group := lifecycle.NewGroup(logger)
	group.Add("gateway", func(ctx context.Context) error {
		service.Start(ctx)
		<-ctx.Done()
		return service.Stop()
	})
	if cfg.GRPCListen != "" {
		control := grpcserver.New(cfg.GRPCListen, func(s *grpc.Server) {
			ft12v1.RegisterGatewayServiceServer(s, gatewaygrpc.New(ctx, service))
		})
		logger.Printf("gateway gRPC control listening on %s", cfg.GRPCListen)
		group.Add("gateway-grpc", control.Run)
	}

	if err := group.Run(ctx); err != nil {
		logger.Printf("gateway stopped with error: %v", err)
		return 1
	}
	logger.Println("gateway stopped")
	return 0
}
