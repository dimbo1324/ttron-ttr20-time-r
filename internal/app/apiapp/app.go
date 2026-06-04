package apiapp

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	grpcclient "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/client"
	httpclient "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/http/client"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/api/http/handlers"
	httpserver "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/http/server"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	platformlogging "github.com/dimbo1324/ttron-ttr20-time-r/internal/platform/logging"
)

func Run(args []string) int {
	cfg, err := config.LoadAPI(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "api config failed: %v\n", err)
		return 1
	}

	logger := platformlogging.New(cfg.LogFile)
	logger.Printf("starting ft12 api (http=%s emulator-grpc=%s gateway-grpc=%s)", cfg.HTTPListen, cfg.EmulatorGRPC, cfg.GatewayGRPC)

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	emulatorClient, emulatorConn, err := grpcclient.DialEmulator(rootCtx, cfg.EmulatorGRPC)
	if err != nil {
		logger.Printf("emulator grpc dial failed: %v", err)
		return 1
	}
	defer emulatorConn.Close()
	gatewayClient, gatewayConn, err := grpcclient.DialGateway(rootCtx, cfg.GatewayGRPC)
	if err != nil {
		logger.Printf("gateway grpc dial failed: %v", err)
		return 1
	}
	defer gatewayConn.Close()

	handler := handlers.New(
		httpclient.NewEmulatorGRPCClient(emulatorClient),
		httpclient.NewGatewayGRPCClient(gatewayClient),
		handlers.Config{
			RequestTimeout: cfg.RequestTimeout,
			EmulatorGRPC:   cfg.EmulatorGRPC,
			GatewayGRPC:    cfg.GatewayGRPC,
		},
	)
	server := httpserver.New(httpserver.Config{
		Address:    cfg.HTTPListen,
		CORSOrigin: cfg.CORSOrigin,
	}, handler.Routes(), logger)

	if err := server.Run(rootCtx); err != nil {
		logger.Printf("api stopped with error: %v", err)
		return 1
	}
	logger.Println("api stopped")
	return 0
}
