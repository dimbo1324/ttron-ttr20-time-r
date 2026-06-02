package emu

import (
	"fmt"
	"log"
	"net"
	"sln/internal/config"
	"sync"
)

// Server представляет TCP-эмулятор
type Server struct {
	cfg    *config.Config
	logger *log.Logger
	ln     net.Listener
	wg     sync.WaitGroup
	close  chan struct{}
	closed bool
	mu     sync.Mutex
}

// NewServer создаёт новый экземпляр сервера с конфигом и логгером
func NewServer(cfg *config.Config, logger *log.Logger) *Server {
	return &Server{
		cfg:    cfg,
		logger: logger,
		close:  make(chan struct{}),
	}
}

// Start запускает TCP-слушатель и принимает входящие подключения
// Функция блокирует до Stop() или ошибки
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.ln = ln
	s.logger.Printf("listening on %s", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			// При закрытии сервера Accept вернёт ошибку; тогда корректно выходим
			select {
			case <-s.close:
				return nil
			default:
				s.logger.Printf("accept error: %v", err)
				continue
			}
		}
		// Новое подключение - обрабатываем в отдельной горутине
		s.logger.Printf("accepted connection from %s", conn.RemoteAddr())
		s.wg.Add(1)
		go func(c net.Conn) {
			defer s.wg.Done()
			handleConnection(c, s.cfg, s.logger)
		}(conn)
	}
}

// Stop корректно останавливает сервер: закрывает listener и ждёт хендлер-горутин
func (s *Server) Stop() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	s.mu.Unlock()

	close(s.close)
	if s.ln != nil {
		_ = s.ln.Close()
	}
	s.logger.Printf("closing server, waiting for handlers...")
	s.wg.Wait()
	s.logger.Printf("server stopped")
}
