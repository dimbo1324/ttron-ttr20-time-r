package dto

import ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"

type EmulatorStatusDTO struct {
	State             string        `json:"state"`
	ListenAddr        string        `json:"listenAddr"`
	ChecksumMode      string        `json:"checksumMode"`
	ActiveConnections int64         `json:"activeConnections"`
	TotalConnections  int64         `json:"totalConnections"`
	TotalRequests     int64         `json:"totalRequests"`
	TotalResponses    int64         `json:"totalResponses"`
	ProtocolErrors    int64         `json:"protocolErrors"`
	LastError         string        `json:"lastError,omitempty"`
	LastRequestTime   *string       `json:"lastRequestTime,omitempty"`
	LastResponseTime  *string       `json:"lastResponseTime,omitempty"`
	FaultMode         *FaultModeDTO `json:"faultMode,omitempty"`
	RecentFramesCount int64         `json:"recentFramesCount"`
}

type FaultModeDTO struct {
	ResponseDelayMs            int64   `json:"responseDelayMs"`
	CorruptChecksum            bool    `json:"corruptChecksum"`
	CorruptChecksumProbability float64 `json:"corruptChecksumProbability"`
	FragmentResponse           bool    `json:"fragmentResponse"`
	FragmentProbability        float64 `json:"fragmentProbability"`
	FragmentDelayMs            int64   `json:"fragmentDelayMs"`
	NoResponse                 bool    `json:"noResponse"`
	CloseAfterRequest          bool    `json:"closeAfterRequest"`
}

func EmulatorStatus(status *ft12v1.EmulatorStatus) EmulatorStatusDTO {
	if status == nil {
		return EmulatorStatusDTO{State: "unavailable", ChecksumMode: "unspecified"}
	}
	return EmulatorStatusDTO{
		State:             ServiceState(status.GetState()),
		ListenAddr:        status.GetListenAddr(),
		ChecksumMode:      ChecksumMode(status.GetChecksumMode()),
		ActiveConnections: int64(status.GetActiveConnections()),
		TotalConnections:  int64(status.GetTotalConnections()),
		TotalRequests:     int64(status.GetTotalRequests()),
		TotalResponses:    int64(status.GetTotalResponses()),
		ProtocolErrors:    int64(status.GetProtocolErrors()),
		LastError:         status.GetLastError(),
		LastRequestTime:   Timestamp(status.GetLastRequestTime()),
		LastResponseTime:  Timestamp(status.GetLastResponseTime()),
		FaultMode:         FaultMode(status.GetFaultMode()),
		RecentFramesCount: int64(status.GetRecentFramesCount()),
	}
}

func FaultMode(fault *ft12v1.FaultMode) *FaultModeDTO {
	if fault == nil {
		return nil
	}
	return &FaultModeDTO{
		ResponseDelayMs:            fault.GetResponseDelayMs(),
		CorruptChecksum:            fault.GetCorruptChecksum(),
		CorruptChecksumProbability: fault.GetCorruptChecksumProbability(),
		FragmentResponse:           fault.GetFragmentResponse(),
		FragmentProbability:        fault.GetFragmentProbability(),
		FragmentDelayMs:            fault.GetFragmentDelayMs(),
		NoResponse:                 fault.GetNoResponse(),
		CloseAfterRequest:          fault.GetCloseAfterRequest(),
	}
}

func FaultModeProto(fault FaultModeDTO) *ft12v1.FaultMode {
	return &ft12v1.FaultMode{
		ResponseDelayMs:            fault.ResponseDelayMs,
		CorruptChecksum:            fault.CorruptChecksum,
		CorruptChecksumProbability: fault.CorruptChecksumProbability,
		FragmentResponse:           fault.FragmentResponse,
		FragmentProbability:        fault.FragmentProbability,
		FragmentDelayMs:            fault.FragmentDelayMs,
		NoResponse:                 fault.NoResponse,
		CloseAfterRequest:          fault.CloseAfterRequest,
	}
}
