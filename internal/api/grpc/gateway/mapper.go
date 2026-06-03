package gateway

import (
	"time"

	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
	domain "github.com/dimbo1324/ttron-ttr20-time-r/internal/gateway"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/observability/events"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapStatus(status domain.Status) *ft12v1.GatewayStatus {
	state := ft12v1.ServiceState_SERVICE_STATE_STOPPED
	if status.Running {
		state = ft12v1.ServiceState_SERVICE_STATE_RUNNING
	}
	if status.LastError != "" && status.Running {
		state = ft12v1.ServiceState_SERVICE_STATE_DEGRADED
	}
	return &ft12v1.GatewayStatus{
		State:                  state,
		TargetAddr:             status.TargetAddress,
		ChecksumMode:           mapChecksumMode(status.ChecksumMode),
		PollingIntervalMs:      int64(status.PollingInterval / time.Millisecond),
		RequestTimeoutMs:       int64(status.RequestTimeout / time.Millisecond),
		ConnectTimeoutMs:       int64(status.ConnectTimeout / time.Millisecond),
		Connected:              status.Connected,
		ConnectionAttempts:     status.ConnectionAttempts,
		SuccessfulReads:        status.SuccessfulReads,
		FailedReads:            status.FailedReads,
		Reconnects:             status.ReconnectCount,
		LastSuccessfulReadTime: mapTime(status.LastSuccessfulReadTime),
		LastDeviceTime:         mapTime(status.LastParsedDeviceTime),
		LastError:              status.LastError,
		LastTxTime:             mapTime(status.LastTXTimestamp),
		LastRxTime:             mapTime(status.LastRXTimestamp),
		RecentFramesCount:      uint32(status.RecentFramesCount),
	}
}

func mapEvents(records []events.FrameRecord, limit uint32) []*ft12v1.FrameEvent {
	if limit > 0 && int(limit) < len(records) {
		records = records[len(records)-int(limit):]
	}
	out := make([]*ft12v1.FrameEvent, 0, len(records))
	for i, record := range records {
		out = append(out, &ft12v1.FrameEvent{
			Id:           uint64(i + 1),
			Timestamp:    mapTime(record.Timestamp),
			Service:      record.Service,
			Direction:    mapDirection(record.Direction),
			RemoteAddr:   record.RemoteAddr,
			ChecksumMode: mapChecksumMode(record.ChecksumMode),
			RawHex:       record.RawHex,
			Command:      record.Command,
			Error:        record.Error,
			Message:      record.Error,
		})
	}
	return out
}

func mapChecksumMode(mode string) ft12v1.ChecksumMode {
	switch mode {
	case "sum":
		return ft12v1.ChecksumMode_CHECKSUM_MODE_SUM
	case "crc16":
		return ft12v1.ChecksumMode_CHECKSUM_MODE_CRC16
	default:
		return ft12v1.ChecksumMode_CHECKSUM_MODE_UNSPECIFIED
	}
}

func mapDirection(direction string) ft12v1.EventDirection {
	switch direction {
	case "RX":
		return ft12v1.EventDirection_EVENT_DIRECTION_RX
	case "TX":
		return ft12v1.EventDirection_EVENT_DIRECTION_TX
	case "ERR":
		return ft12v1.EventDirection_EVENT_DIRECTION_ERROR
	default:
		return ft12v1.EventDirection_EVENT_DIRECTION_SYSTEM
	}
}

func mapTime(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}
