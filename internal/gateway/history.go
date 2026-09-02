package gateway

import (
	"fmt"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/clock"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/health"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/observability/events"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/command"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/util"
)

func (s *Service) recordTX(remote string, raw []byte, cmd string) {
	now := time.Now()
	s.history.Add(events.FrameRecord{Timestamp: now, Direction: events.DirectionTX, Service: events.ServiceGateway, RemoteAddr: remote, RawHex: util.HexDump(raw), Command: cmd, ChecksumMode: string(s.mode)})
	s.recordTXTimestamp(now)
}

func (s *Service) recordRX(remote string, raw []byte, cmd string) {
	now := time.Now()
	s.history.Add(events.FrameRecord{Timestamp: now, Direction: events.DirectionRX, Service: events.ServiceGateway, RemoteAddr: remote, RawHex: util.HexDump(raw), Command: cmd, ChecksumMode: string(s.mode)})
	s.recordRXTimestamp(now)
}

func (s *Service) recordProtocolError(remote string, err error) {
	s.history.Add(events.FrameRecord{Timestamp: time.Now(), Direction: events.DirectionError, Service: events.ServiceGateway, RemoteAddr: remote, ChecksumMode: string(s.mode), Error: err.Error(), Message: err.Error()})
}

func (s *Service) recordSystemEvent(command, message string, at time.Time) {
	if at.IsZero() {
		at = time.Now()
	}
	s.history.Add(events.FrameRecord{
		Timestamp:    at,
		Direction:    events.DirectionSystem,
		Service:      events.ServiceGateway,
		RemoteAddr:   s.cfg.Target,
		ChecksumMode: string(s.mode),
		Command:      command,
		Message:      message,
	})
}

func (s *Service) recordClockTransition(report clock.Report) {
	s.recordSystemEvent("clock-state", fmt.Sprintf(
		"clock state %s -> %s skew=%s median=%s drift=%s/day warn=%s critical=%s samples=%d",
		report.PreviousState, report.State, report.Skew, report.MedianSkew,
		report.Drift.PerDay, report.Thresholds.Warn, report.Thresholds.Critical, report.SampleCount,
	), report.At)
}

func (s *Service) recordHealthTransition(transition health.Transition) {
	s.recordSystemEvent("device-state", fmt.Sprintf(
		"device state %s -> %s reason=%s", transition.From, transition.To, transition.Reason,
	), transition.At)
}

func (s *Service) recordIdentityEvent(identity command.Identity, at time.Time) {
	s.recordSystemEvent("read-identity", fmt.Sprintf(
		"device identity model=%s serial=%s firmware=%s", identity.Model, identity.Serial, identity.Firmware,
	), at)
}
