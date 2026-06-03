package clientapp

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/client"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	platformlogging "github.com/dimbo1324/ttron-ttr20-time-r/internal/platform/logging"
)

func Run(args []string) int {
	cfg, err := config.LoadClient(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "client config failed: %v\n", err)
		return 1
	}

	logger := platformlogging.New(cfg.LogFile)
	logger.Printf("starting ft12 client (target=%s:%d adapter=%d crc=%s timeout=%dms retries=%d)",
		cfg.Host, cfg.Port, cfg.AdapterAddr, cfg.CRCMode, cfg.TimeoutMs, cfg.Retries)

	cl := client.New(cfg, logger)
	if err := cl.Start(); err != nil {
		logger.Printf("client start failed: %v", err)
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	logger.Printf("signal received, stopping client...")
	cl.Stop()
	logger.Println("client stopped")
	return 0
}
