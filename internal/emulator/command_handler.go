package emulator

import (
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/command"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
)

func (s *Service) BuildResponse(req frame.Frame) ([]byte, string, bool, error) {
	data := req.DataBytes()
	id, err := command.ParseID(data)
	if err != nil {
		resp, buildErr := s.buildAckResponse(req, data)
		return resp, "ack", false, buildErr
	}

	switch id {
	case command.ReadTime:
		if err := command.ParseReadTimeRequest(data); err != nil {
			break
		}
		resp, err := s.buildReadTimeResponse(req)
		return resp, command.NameReadTime, true, err
	case command.ReadIdentity:
		if err := command.ParseReadIdentityRequest(data); err != nil {
			break
		}
		resp, err := s.buildReadIdentityResponse(req)
		return resp, command.NameReadIdentity, true, err
	}

	resp, err := s.buildAckResponse(req, data)
	return resp, s.commandName(data), false, err
}

func (s *Service) commandName(data []byte) string {
	id, err := command.ParseID(data)
	if err != nil {
		return "ack"
	}
	return s.commands.Name(id)
}
