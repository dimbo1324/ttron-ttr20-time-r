package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/client"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/logging"
)

func main() {
	cfg := config.LoadClient()
	logger := logging.New(cfg.LogFile)
	logger.Printf("starting ft12 client (target=%s:%d adapter=%d crc=%s timeout=%dms retries=%d)",
		cfg.Host, cfg.Port, cfg.AdapterAddr, cfg.CRCMode, cfg.TimeoutMs, cfg.Retries)

	cl := client.New(cfg, logger)
	if err := cl.Start(); err != nil {
		logger.Fatalf("client start failed: %v", err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	logger.Printf("signal received, stopping client...")
	cl.Stop()
	logger.Println("client stopped")
}
