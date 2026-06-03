package emulator

import (
	"context"

	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
	domain "github.com/dimbo1324/ttron-ttr20-time-r/internal/emulator"
)

type Service struct {
	ft12v1.UnimplementedEmulatorServiceServer
	emulator *domain.Service
}

func New(service *domain.Service) *Service {
	return &Service{emulator: service}
}

func (s *Service) GetStatus(context.Context, *ft12v1.GetEmulatorStatusRequest) (*ft12v1.GetEmulatorStatusResponse, error) {
	return &ft12v1.GetEmulatorStatusResponse{Status: mapStatus(s.emulator.Status())}, nil
}

func (s *Service) GetFaultMode(context.Context, *ft12v1.GetFaultModeRequest) (*ft12v1.GetFaultModeResponse, error) {
	return &ft12v1.GetFaultModeResponse{FaultMode: mapFaultMode(s.emulator.FaultMode())}, nil
}

func (s *Service) SetFaultMode(_ context.Context, req *ft12v1.SetFaultModeRequest) (*ft12v1.SetFaultModeResponse, error) {
	fault := s.emulator.SetFaultMode(faultFromProto(req.GetFaultMode()))
	return &ft12v1.SetFaultModeResponse{
		FaultMode: mapFaultMode(fault),
		Status:    mapStatus(s.emulator.Status()),
	}, nil
}

func (s *Service) GetRecentEvents(_ context.Context, req *ft12v1.GetRecentEventsRequest) (*ft12v1.GetRecentEventsResponse, error) {
	return &ft12v1.GetRecentEventsResponse{Events: mapEvents(s.emulator.Snapshot().Recent, req.GetLimit())}, nil
}
