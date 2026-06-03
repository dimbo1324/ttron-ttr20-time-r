package emulator

import (
	"time"

	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
	domain "github.com/dimbo1324/ttron-ttr20-time-r/internal/emulator"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/observability/events"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapStatus(status domain.Status) *ft12v1.EmulatorStatus {
	state := ft12v1.ServiceState_SERVICE_STATE_STOPPED
	if status.Running {
		state = ft12v1.ServiceState_SERVICE_STATE_RUNNING
	}
	if status.LastError != "" && status.Running {
		state = ft12v1.ServiceState_SERVICE_STATE_DEGRADED
	}
	return &ft12v1.EmulatorStatus{
		State:             state,
		ListenAddr:        status.ListenAddress,
		ChecksumMode:      mapChecksumMode(status.ChecksumMode),
		ActiveConnections: uint32(status.ActiveConnections),
		TotalConnections:  status.TotalConnections,
		TotalRequests:     status.TotalRequests,
		TotalResponses:    status.TotalResponses,
		ProtocolErrors:    status.TotalProtocolErrors,
		LastError:         status.LastError,
		LastRequestTime:   mapTime(status.LastRequestTime),
		LastResponseTime:  mapTime(status.LastResponseTime),
		FaultMode:         mapFaultMode(status.FaultMode),
		RecentFramesCount: uint32(status.RecentFramesCount),
	}
}

func mapFaultMode(fault domain.FaultMode) *ft12v1.FaultMode {
	return &ft12v1.FaultMode{
		ResponseDelayMs:            int64(fault.ResponseDelay / time.Millisecond),
		CorruptChecksum:            fault.BadChecksumProb > 0,
		CorruptChecksumProbability: fault.BadChecksumProb,
		FragmentResponse:           fault.FragmentProb > 0,
		FragmentProbability:        fault.FragmentProb,
		FragmentDelayMs:            int64(fault.FragmentDelay / time.Millisecond),
		NoResponse:                 fault.NoResponse,
		CloseAfterRequest:          fault.CloseAfterRequest,
	}
}

func faultFromProto(fault *ft12v1.FaultMode) domain.FaultMode {
	if fault == nil {
		return domain.FaultMode{}
	}
	corruptProb := fault.CorruptChecksumProbability
	if fault.CorruptChecksum && corruptProb == 0 {
		corruptProb = 1
	}
	fragmentProb := fault.FragmentProbability
	if fault.FragmentResponse && fragmentProb == 0 {
		fragmentProb = 1
	}
	return domain.FaultMode{
		ResponseDelay:     time.Duration(fault.ResponseDelayMs) * time.Millisecond,
		BadChecksumProb:   corruptProb,
		FragmentProb:      fragmentProb,
		FragmentDelay:     time.Duration(fault.FragmentDelayMs) * time.Millisecond,
		NoResponse:        fault.NoResponse,
		CloseAfterRequest: fault.CloseAfterRequest,
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
