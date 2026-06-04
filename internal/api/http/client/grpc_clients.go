package client

import (
	"context"

	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
)

type EmulatorClient interface {
	GetStatus(context.Context) (*ft12v1.EmulatorStatus, error)
	GetFaultMode(context.Context) (*ft12v1.FaultMode, error)
	SetFaultMode(context.Context, *ft12v1.FaultMode) (*ft12v1.FaultMode, *ft12v1.EmulatorStatus, error)
	GetRecentEvents(context.Context, uint32) ([]*ft12v1.FrameEvent, error)
}

type GatewayClient interface {
	GetStatus(context.Context) (*ft12v1.GatewayStatus, error)
	StartPolling(context.Context) (*ft12v1.GatewayStatus, error)
	StopPolling(context.Context) (*ft12v1.GatewayStatus, error)
	GetLastReadTime(context.Context) (*ft12v1.GetLastReadTimeResponse, error)
	GetRecentEvents(context.Context, uint32) ([]*ft12v1.FrameEvent, error)
}

type EmulatorGRPCClient struct {
	client ft12v1.EmulatorServiceClient
}

func NewEmulatorGRPCClient(client ft12v1.EmulatorServiceClient) *EmulatorGRPCClient {
	return &EmulatorGRPCClient{client: client}
}

func (c *EmulatorGRPCClient) GetStatus(ctx context.Context) (*ft12v1.EmulatorStatus, error) {
	resp, err := c.client.GetStatus(ctx, &ft12v1.GetEmulatorStatusRequest{})
	if err != nil {
		return nil, err
	}
	return resp.GetStatus(), nil
}

func (c *EmulatorGRPCClient) GetFaultMode(ctx context.Context) (*ft12v1.FaultMode, error) {
	resp, err := c.client.GetFaultMode(ctx, &ft12v1.GetFaultModeRequest{})
	if err != nil {
		return nil, err
	}
	return resp.GetFaultMode(), nil
}

func (c *EmulatorGRPCClient) SetFaultMode(ctx context.Context, fault *ft12v1.FaultMode) (*ft12v1.FaultMode, *ft12v1.EmulatorStatus, error) {
	resp, err := c.client.SetFaultMode(ctx, &ft12v1.SetFaultModeRequest{FaultMode: fault})
	if err != nil {
		return nil, nil, err
	}
	return resp.GetFaultMode(), resp.GetStatus(), nil
}

func (c *EmulatorGRPCClient) GetRecentEvents(ctx context.Context, limit uint32) ([]*ft12v1.FrameEvent, error) {
	resp, err := c.client.GetRecentEvents(ctx, &ft12v1.GetRecentEventsRequest{Limit: limit})
	if err != nil {
		return nil, err
	}
	return resp.GetEvents(), nil
}

type GatewayGRPCClient struct {
	client ft12v1.GatewayServiceClient
}

func NewGatewayGRPCClient(client ft12v1.GatewayServiceClient) *GatewayGRPCClient {
	return &GatewayGRPCClient{client: client}
}

func (c *GatewayGRPCClient) GetStatus(ctx context.Context) (*ft12v1.GatewayStatus, error) {
	resp, err := c.client.GetStatus(ctx, &ft12v1.GetGatewayStatusRequest{})
	if err != nil {
		return nil, err
	}
	return resp.GetStatus(), nil
}

func (c *GatewayGRPCClient) StartPolling(ctx context.Context) (*ft12v1.GatewayStatus, error) {
	resp, err := c.client.StartPolling(ctx, &ft12v1.StartPollingRequest{})
	if err != nil {
		return nil, err
	}
	return resp.GetStatus(), nil
}

func (c *GatewayGRPCClient) StopPolling(ctx context.Context) (*ft12v1.GatewayStatus, error) {
	resp, err := c.client.StopPolling(ctx, &ft12v1.StopPollingRequest{})
	if err != nil {
		return nil, err
	}
	return resp.GetStatus(), nil
}

func (c *GatewayGRPCClient) GetLastReadTime(ctx context.Context) (*ft12v1.GetLastReadTimeResponse, error) {
	return c.client.GetLastReadTime(ctx, &ft12v1.GetLastReadTimeRequest{})
}

func (c *GatewayGRPCClient) GetRecentEvents(ctx context.Context, limit uint32) ([]*ft12v1.FrameEvent, error) {
	resp, err := c.client.GetRecentEvents(ctx, &ft12v1.GetRecentEventsRequest{Limit: limit})
	if err != nil {
		return nil, err
	}
	return resp.GetEvents(), nil
}
