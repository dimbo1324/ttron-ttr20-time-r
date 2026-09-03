package gatewayapp

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
	gatewaygrpc "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/gateway"
	grpcserver "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/server"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/devices"
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
	logger.Printf("starting ft12 gateway (target=%s crc=%s schedule=%s interval=%s offset=%s timeout=%s retries=%d grpc=%s devices=%q)",
		cfg.Target, cfg.CRCMode, cfg.ScheduleMode, cfg.PollInterval, cfg.PollOffset,
		cfg.RequestTimeout, cfg.RetryAttempts, cfg.GRPCListen, cfg.DevicesFile)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	group := lifecycle.NewGroup(logger)

	service, supervisor, err := buildRuntime(ctx, cfg, group, logger)
	if err != nil {
		logger.Printf("gateway runtime creation failed: %v", err)
		return 1
	}

	if cfg.GRPCListen != "" && service != nil {
		control := grpcserver.New(cfg.GRPCListen, func(s *grpc.Server) {
			api := gatewaygrpc.New(ctx, service)
			// Without an inventory there is no supervisor, and the control
			// plane reports a fleet of one rather than nothing at all.
			if supervisor != nil {
				api = api.WithFleet(supervisor)
			}
			ft12v1.RegisterGatewayServiceServer(s, api)
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

// buildRuntime returns the service the control plane drives and, in inventory
// mode, the supervisor behind it. Both are nil-able: an inventory with nothing
// enabled has no primary device to control.
func buildRuntime(ctx context.Context, cfg *config.GatewayConfig, group *lifecycle.Group, logger *log.Logger) (*gateway.Service, *gateway.Supervisor, error) {
	if cfg.DevicesFile == "" {
		service, err := gateway.NewService(cfg, logger)
		if err != nil {
			return nil, nil, err
		}
		group.Add("gateway", func(ctx context.Context) error {
			service.Start(ctx)
			<-ctx.Done()
			return service.Stop()
		})
		return service, nil, nil
	}

	registry, err := devices.Load(cfg.DevicesFile)
	if err != nil {
		return nil, nil, err
	}
	supervisor, err := gateway.NewSupervisor(cfg, registry, logger)
	if err != nil {
		return nil, nil, err
	}
	logger.Printf("gateway device inventory loaded file=%s devices=%d enabled=%d",
		cfg.DevicesFile, registry.Len(), len(registry.Enabled()))
	group.Add("gateway-supervisor", supervisor.Run)

	primary, ok := supervisor.Primary()
	if !ok {
		if cfg.GRPCListen != "" {
			logger.Printf("gateway gRPC control disabled: device inventory has no enabled devices")
		}
		return nil, supervisor, nil
	}
	if cfg.GRPCListen != "" {
		logger.Printf("gateway gRPC control bound to primary device=%s", primary.DeviceID())
	}
	return primary, supervisor, nil
}
