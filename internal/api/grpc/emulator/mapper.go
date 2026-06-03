package emulator

import (
	"time"

	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/mapping"
	domain "github.com/dimbo1324/ttron-ttr20-time-r/internal/emulator"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/observability/events"
)

func mapStatus(status domain.Status) *ft12v1.EmulatorStatus {
	return &ft12v1.EmulatorStatus{
		State:             mapping.ServiceState(status.Running, status.LastError),
		ListenAddr:        status.ListenAddress,
		ChecksumMode:      mapping.ChecksumMode(status.ChecksumMode),
		ActiveConnections: uint32(status.ActiveConnections),
		TotalConnections:  status.TotalConnections,
		TotalRequests:     status.TotalRequests,
		TotalResponses:    status.TotalResponses,
		ProtocolErrors:    status.TotalProtocolErrors,
		LastError:         status.LastError,
		LastRequestTime:   mapping.Time(status.LastRequestTime),
		LastResponseTime:  mapping.Time(status.LastResponseTime),
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
	return mapping.Events(records, limit)
}
