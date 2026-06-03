package gateway

import (
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/command"
)

func (s *Service) recordTXTimestamp(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.LastTXTimestamp = now
}

func (s *Service) recordRXTimestamp(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.LastRXTimestamp = now
}

func (s *Service) recordSuccess(parsed command.ReadTimeResponse) {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.SuccessfulReads++
	s.status.LastSuccessfulReadTime = now
	s.status.LastParsedDeviceTime = parsed.Time
	s.status.LastError = ""
}

func (s *Service) recordFailure(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.FailedReads++
	if err != nil {
		s.status.LastError = err.Error()
	}
}

func (s *Service) incrementConnectionAttempts() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.ConnectionAttempts++
}

func (s *Service) incrementReconnects() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.ReconnectCount++
}

func (s *Service) setConnected(connected bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.Connected = connected
}

func (s *Service) setRunning(running bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.Running = running
}
