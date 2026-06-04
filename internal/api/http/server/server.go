package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/api/http/metrics"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/api/http/middleware"
)

type Logger interface {
	Printf(format string, v ...any)
}

type Config struct {
	Address      string
	CORSOrigin   string
	Metrics      *metrics.Registry
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type Server struct {
	cfg    Config
	logger Logger
	server *http.Server

	mu sync.RWMutex
	ln net.Listener
}

func New(cfg Config, handler http.Handler, logger Logger) *Server {
	if cfg.ReadTimeout <= 0 {
		cfg.ReadTimeout = 5 * time.Second
	}
	if cfg.WriteTimeout <= 0 {
		cfg.WriteTimeout = 10 * time.Second
	}
	if cfg.IdleTimeout <= 0 {
		cfg.IdleTimeout = 60 * time.Second
	}
	wrapped := middleware.Chain(
		handler,
		middleware.RequestID(),
		middleware.Recovery(logger),
		middleware.CORS(cfg.CORSOrigin),
		middleware.Metrics(func(method, path string, status int, elapsed time.Duration) {
			if cfg.Metrics != nil {
				cfg.Metrics.Observe(method, path, status, elapsed)
			}
		}),
		middleware.Logging(logger),
	)
	return &Server{
		cfg:    cfg,
		logger: logger,
		server: &http.Server{
			Addr:         cfg.Address,
			Handler:      wrapped,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}
}

func (s *Server) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.cfg.Address)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.ln = ln
	s.mu.Unlock()
	if s.logger != nil {
		s.logger.Printf("http api listening on %s", ln.Addr().String())
	}

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = s.server.Shutdown(shutdownCtx)
		case <-done:
		}
	}()

	err = s.server.Serve(ln)
	close(done)
	if ctx.Err() != nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Addr() net.Addr {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ln == nil {
		return nil
	}
	return s.ln.Addr()
}
