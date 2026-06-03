package emulator

import (
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/observability/events"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/util"
)

func (s *Service) recordFrame(direction events.Direction, remote, rawHex, cmd, errText string) {
	s.history.Add(events.FrameRecord{
		Timestamp:    s.now(),
		Direction:    direction,
		Service:      events.ServiceEmulator,
		RemoteAddr:   remote,
		RawHex:       rawHex,
		Command:      cmd,
		ChecksumMode: string(s.mode),
		Error:        errText,
		Message:      errText,
	})
}

func (s *Service) recordRX(remote string, req frame.Frame) {
	s.recordFrame(events.DirectionRX, remote, util.HexDump(req.RawBytes()), commandName(req.DataBytes()), "")
	s.recordRequest()
}

func (s *Service) recordTX(remote string, raw []byte, cmd string) {
	s.recordFrame(events.DirectionTX, remote, util.HexDump(raw), cmd, "")
	s.recordResponse()
}

func (s *Service) recordError(remote string, err error) {
	errText := ""
	if err != nil {
		errText = err.Error()
	}
	s.recordFrame(events.DirectionError, remote, "", "", errText)
	s.recordProtocolError(errText)
}
