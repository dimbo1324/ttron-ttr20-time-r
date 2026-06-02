package main

import (
	"os"
	"os/signal"
	"sln/internal/config"
	"sln/internal/emu"
	"sln/internal/logging"
	"syscall"
)

// Точка входа сервера. Загружает конфиг, запускает эмулятор и
// обрабатывает корректное завершение по сигналам
func main() {
	cfg := config.Load()

	logger := logging.New(cfg.LogFile)

	logger.Printf("starting ttp20 emulator (host=%s port=%d crc=%s adapter=%d)",
		cfg.Host, cfg.Port, cfg.CRCMode, cfg.AdapterAddr)

	// Создаём сервер-эмулятор
	srv := emu.NewServer(cfg, logger)

	// Запускаем сервер в отдельной горутине и отслеживаем ошибку
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	// Ловим сигналы для graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	select {
	case sigv := <-sig:
		// При получении сигнала корректно останавливаем сервер
		logger.Printf("received signal %v, shutting down...", sigv)
		srv.Stop()
	case err := <-errCh:
		// Если сервер завершился сам - логируем причину
		if err != nil {
			logger.Printf("server stopped with error: %v", err)
		} else {
			logger.Printf("server stopped")
		}
	}

	logger.Println("bye")
}
