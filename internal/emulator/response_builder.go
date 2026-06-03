package emulator

import "github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"

func (s *Service) buildReadTimeResponse(req frame.Frame) ([]byte, error) {
	return s.wire.EncodeReadTimeResponse(req, s.now())
}

func (s *Service) buildAckResponse(req frame.Frame, data []byte) ([]byte, error) {
	return s.wire.EncodeACK(req, data)
}
