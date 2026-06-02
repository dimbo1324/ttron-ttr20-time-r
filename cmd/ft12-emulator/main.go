package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/emulator"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/logging"
)

func main() {
	cfg := config.LoadEmulator()
	logger := logging.New(cfg.LogFile)
	logger.Printf("starting ft12 emulator (host=%s port=%d crc=%s adapter=%d)",
		cfg.Host, cfg.Port, cfg.CRCMode, cfg.AdapterAddr)

	srv := emulator.NewServer(cfg, logger)
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	select {
	case sigv := <-sig:
		logger.Printf("received signal %v, shutting down...", sigv)
		srv.Stop()
	case err := <-errCh:
		if err != nil {
			logger.Printf("server stopped with error: %v", err)
		} else {
			logger.Printf("server stopped")
		}
	}
	logger.Println("bye")
}
