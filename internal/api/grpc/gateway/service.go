package gateway

import (
	"context"

	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/mapping"
	domain "github.com/dimbo1324/ttron-ttr20-time-r/internal/gateway"
)

// FleetSource is anything that can report on more than one device -- in
// practice the supervisor. It is an interface, and optional, because a gateway
// configured without a device inventory has no supervisor at all; see GetFleet
// for what happens then.
type FleetSource interface {
	Fleet() domain.FleetStatus
}

type Service struct {
	ft12v1.UnimplementedGatewayServiceServer
	gateway *domain.Service
	fleet   FleetSource
	rootCtx context.Context
}

func New(rootCtx context.Context, service *domain.Service) *Service {
	return &Service{rootCtx: rootCtx, gateway: service}
}

// WithFleet attaches the multi-device view. Without it the service still
// answers GetFleet, using the single device it controls.
func (s *Service) WithFleet(fleet FleetSource) *Service {
	s.fleet = fleet
	return s
}

func (s *Service) GetStatus(context.Context, *ft12v1.GetGatewayStatusRequest) (*ft12v1.GetGatewayStatusResponse, error) {
	return &ft12v1.GetGatewayStatusResponse{Status: mapStatus(s.gateway.Status())}, nil
}

func (s *Service) StartPolling(context.Context, *ft12v1.StartPollingRequest) (*ft12v1.StartPollingResponse, error) {
	s.gateway.Start(s.rootCtx)
	return &ft12v1.StartPollingResponse{Status: mapStatus(s.gateway.Status())}, nil
}

func (s *Service) StopPolling(context.Context, *ft12v1.StopPollingRequest) (*ft12v1.StopPollingResponse, error) {
	if err := s.gateway.Stop(); err != nil {
		return nil, err
	}
	return &ft12v1.StopPollingResponse{Status: mapStatus(s.gateway.Status())}, nil
}

func (s *Service) GetRecentEvents(_ context.Context, req *ft12v1.GetRecentEventsRequest) (*ft12v1.GetRecentEventsResponse, error) {
	return &ft12v1.GetRecentEventsResponse{Events: mapEvents(s.gateway.Snapshot().Recent, req.GetLimit())}, nil
}

func (s *Service) GetLastReadTime(context.Context, *ft12v1.GetLastReadTimeRequest) (*ft12v1.GetLastReadTimeResponse, error) {
	status := s.gateway.Status()
	return &ft12v1.GetLastReadTimeResponse{
		DeviceTime: mapping.Time(status.LastParsedDeviceTime),
		ReadTime:   mapping.Time(status.LastSuccessfulReadTime),
		Available:  !status.LastParsedDeviceTime.IsZero(),
	}, nil
}

// GetFleet answers for every device, and for a gateway without an inventory
// that is a fleet of one. Callers therefore never have to ask which mode the
// gateway is in before reading the reply.
func (s *Service) GetFleet(context.Context, *ft12v1.GetFleetRequest) (*ft12v1.GetFleetResponse, error) {
	if s.fleet != nil {
		return mapFleet(s.fleet.Fleet()), nil
	}
	return mapFleet(domain.SummarizeFleet([]domain.Status{s.gateway.Status()})), nil
}
