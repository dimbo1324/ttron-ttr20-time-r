package emulator

import (
	"fmt"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/command"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
)

func (s *Service) BuildResponse(req frame.Frame) ([]byte, string, bool, error) {
	data := req.DataBytes()
	if err := command.ParseReadTimeRequest(data); err == nil {
		resp, err := s.buildReadTimeResponse(req)
		return resp, "read-time", true, err
	}

	resp, err := s.buildAckResponse(req, data)
	cmd := commandName(data)
	if cmd == "" {
		cmd = "ack"
	}
	return resp, cmd, false, err
}

func commandName(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	if data[0] == byte(command.ReadTime) {
		return "read-time"
	}
	return fmt.Sprintf("unknown-0x%02X", data[0])
}
