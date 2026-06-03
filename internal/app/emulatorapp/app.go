package emulatorapp

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	emulatorgrpc "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/emulator"
	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
	grpcserver "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/server"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/emulator"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/platform/lifecycle"
	platformlogging "github.com/dimbo1324/ttron-ttr20-time-r/internal/platform/logging"
	"google.golang.org/grpc"
)

func Run(args []string) int {
	cfg, err := config.LoadEmulator(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "emulator config failed: %v\n", err)
		return 1
	}

	logger := platformlogging.New(cfg.LogFile)
	logger.Printf("starting ft12 emulator (listen=%s crc=%s adapter=%d grpc=%s)",
		cfg.ListenAddress(), cfg.CRCMode, cfg.AdapterAddr, cfg.GRPCListen)

	service, err := emulator.NewService(cfg, logger)
	if err != nil {
		logger.Printf("emulator service creation failed: %v", err)
		return 1
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	group := lifecycle.NewGroup(logger)
	group.Add("emulator", service.Run)
	if cfg.GRPCListen != "" {
		control := grpcserver.New(cfg.GRPCListen, func(s *grpc.Server) {
			ft12v1.RegisterEmulatorServiceServer(s, emulatorgrpc.New(service))
		})
		logger.Printf("emulator gRPC control listening on %s", cfg.GRPCListen)
		group.Add("emulator-grpc", control.Run)
	}

	if err := group.Run(ctx); err != nil {
		logger.Printf("emulator stopped with error: %v", err)
		return 1
	}
	logger.Println("bye")
	return 0
}
