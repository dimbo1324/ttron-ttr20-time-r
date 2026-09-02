package gateway

import (
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/clock"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/health"
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

func (s *Service) recordRoundTrip(rtt time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.LastRoundTrip = rtt
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

func (s *Service) incrementRetries() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.Retry.TotalRetries++
}

func (s *Service) incrementExhaustedPolls() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.Retry.ExhaustedPolls++
}

func (s *Service) incrementProtocolErrors() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.ProtocolErrors++
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

func (s *Service) recordIdentity(identity command.Identity, at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.Identity = IdentityStatus{
		Known:     true,
		Supported: true,
		Model:     identity.Model,
		Serial:    identity.Serial,
		Firmware:  identity.Firmware,
		ReadAt:    at,
	}
}

func (s *Service) markIdentityUnsupported() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.Identity.Supported = false
}

func (s *Service) identityKnown() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status.Identity.Known
}

func (s *Service) identitySupported() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status.Identity.Supported
}

func (s *Service) observeClock(sample clock.Sample) {
	report, err := s.skew.Observe(sample)
	if err != nil {
		s.logger.Printf("gateway clock sample rejected: %v", err)
		return
	}

	s.logger.Printf("gateway clock skew=%s median=%s drift=%s/day state=%s samples=%d",
		report.Skew, report.MedianSkew, report.Drift.PerDay, report.State, report.SampleCount)

	if report.Transitioned {
		s.logger.Printf("gateway clock state changed from=%s to=%s skew=%s warn=%s critical=%s",
			report.PreviousState, report.State, report.Skew, report.Thresholds.Warn, report.Thresholds.Critical)
		s.recordClockTransition(report)
	}
}

func (s *Service) observeHealthSuccess(latency time.Duration) {
	s.logTransition(s.health.RecordSuccess(latency))
}

func (s *Service) observeHealthFailure(err error) {
	s.logTransition(s.health.RecordFailure(err))
}

func (s *Service) logTransition(transition health.Transition) {
	if !transition.Changed {
		return
	}
	s.logger.Printf("gateway device state changed from=%s to=%s reason=%s", transition.From, transition.To, transition.Reason)
	s.recordHealthTransition(transition)
}
