package gateway

import (
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/observability/events"
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
