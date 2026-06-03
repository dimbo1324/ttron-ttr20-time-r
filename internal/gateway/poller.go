package gateway

import (
	"context"
	"fmt"
	"time"

	transporttcp "github.com/dimbo1324/ttron-ttr20-time-r/internal/transport/tcp"
)

func (s *Service) Run(ctx context.Context) error {
	s.setRunning(true)
	defer s.setRunning(false)

	backoff := NewBackoff(s.cfg.BackoffInitial, s.cfg.BackoffMax)
	for {
		if ctx.Err() != nil {
			return nil
		}
		s.incrementConnectionAttempts()
		s.logger.Printf("gateway connecting target=%s mode=%s", s.cfg.Target, s.mode)
		conn, err := transporttcp.Dial(ctx, transporttcp.ClientConfig{
			Address:        s.cfg.Target,
			ConnectTimeout: s.cfg.ConnectTimeout,
		})
		if err != nil {
			s.recordFailure(fmt.Errorf("connect failed: %w", err))
			delay := backoff.Next()
			if !sleepContext(ctx, delay) {
				return nil
			}
			continue
		}

		backoff.Reset()
		s.setConnected(true)
		s.logger.Printf("gateway connected target=%s", s.cfg.Target)
		err = s.runSession(ctx, conn)
		_ = conn.Close()
		s.setConnected(false)
		if ctx.Err() != nil {
			return nil
		}
		if err != nil {
			s.recordFailure(err)
			s.incrementReconnects()
			s.logger.Printf("gateway session error: %v", err)
		}
		delay := backoff.Next()
		if !sleepContext(ctx, delay) {
			return nil
		}
	}
}

func sleepContext(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		return ctx.Err() == nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
