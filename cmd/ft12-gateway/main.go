package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/gateway"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/logging"
)

func main() {
	cfg := config.LoadGateway()
	logger := logging.New(cfg.LogFile)
	logger.Printf("starting ft12 gateway (target=%s crc=%s interval=%s timeout=%s)",
		cfg.Target, cfg.CRCMode, cfg.PollInterval, cfg.RequestTimeout)

	service, err := gateway.NewService(cfg, logger)
	if err != nil {
		logger.Fatalf("gateway config failed: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := service.Run(ctx); err != nil {
		logger.Fatalf("gateway stopped with error: %v", err)
	}
	logger.Println("gateway stopped")
}
