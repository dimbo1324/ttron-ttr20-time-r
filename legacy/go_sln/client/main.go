package main

import (
	"os"
	"os/signal"
	"sln/client/internal/client"
	"sln/client/internal/config"
	"sln/client/internal/logging"
	"syscall"
)

// Точка входа клиента. Загружает конфиг, запускает опрашивающий клиент
// и организует корректную остановку по сигналу
func main() {
	cfg := config.Load()
	logger := logging.New(cfg.LogFile)

	logger.Printf("starting ttp20 client (target=%s:%d adapter=%d crc=%s timeout=%dms retries=%d)",
		cfg.Host, cfg.Port, cfg.AdapterAddr, cfg.CRCMode, cfg.TimeoutMs, cfg.Retries)

	cl := client.NewClient(cfg, logger)

	// Запускаем опрос в фоне
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
