package gateway

import (
	"time"

	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/mapping"
	domain "github.com/dimbo1324/ttron-ttr20-time-r/internal/gateway"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/observability/events"
)

func mapStatus(status domain.Status) *ft12v1.GatewayStatus {
	return &ft12v1.GatewayStatus{
		State:                  mapping.ServiceState(status.Running, status.LastError),
		TargetAddr:             status.TargetAddress,
		ChecksumMode:           mapping.ChecksumMode(status.ChecksumMode),
		PollingIntervalMs:      int64(status.PollingInterval / time.Millisecond),
		RequestTimeoutMs:       int64(status.RequestTimeout / time.Millisecond),
		ConnectTimeoutMs:       int64(status.ConnectTimeout / time.Millisecond),
		Connected:              status.Connected,
		ConnectionAttempts:     status.ConnectionAttempts,
		SuccessfulReads:        status.SuccessfulReads,
		FailedReads:            status.FailedReads,
		Reconnects:             status.ReconnectCount,
		LastSuccessfulReadTime: mapping.Time(status.LastSuccessfulReadTime),
		LastDeviceTime:         mapping.Time(status.LastParsedDeviceTime),
		LastError:              status.LastError,
		LastTxTime:             mapping.Time(status.LastTXTimestamp),
		LastRxTime:             mapping.Time(status.LastRXTimestamp),
		RecentFramesCount:      uint32(status.RecentFramesCount),
	}
}

func mapEvents(records []events.FrameRecord, limit uint32) []*ft12v1.FrameEvent {
	return mapping.Events(records, limit)
}
