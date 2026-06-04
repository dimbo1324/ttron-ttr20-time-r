package dto

import ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"

type GatewayStatusDTO struct {
	State                  string  `json:"state"`
	TargetAddr             string  `json:"targetAddr"`
	ChecksumMode           string  `json:"checksumMode"`
	PollingIntervalMs      int64   `json:"pollingIntervalMs"`
	RequestTimeoutMs       int64   `json:"requestTimeoutMs"`
	ConnectTimeoutMs       int64   `json:"connectTimeoutMs"`
	Connected              bool    `json:"connected"`
	ConnectionAttempts     int64   `json:"connectionAttempts"`
	SuccessfulReads        int64   `json:"successfulReads"`
	FailedReads            int64   `json:"failedReads"`
	Reconnects             int64   `json:"reconnects"`
	LastSuccessfulReadTime *string `json:"lastSuccessfulReadTime,omitempty"`
	LastDeviceTime         *string `json:"lastDeviceTime,omitempty"`
	LastError              string  `json:"lastError,omitempty"`
	LastTxTime             *string `json:"lastTxTime,omitempty"`
	LastRxTime             *string `json:"lastRxTime,omitempty"`
	RecentFramesCount      int64   `json:"recentFramesCount"`
}

type LastReadTimeDTO struct {
	Available  bool    `json:"available"`
	DeviceTime *string `json:"deviceTime,omitempty"`
	ReadTime   *string `json:"readTime,omitempty"`
}

func GatewayStatus(status *ft12v1.GatewayStatus) GatewayStatusDTO {
	if status == nil {
		return GatewayStatusDTO{State: "unavailable", ChecksumMode: "unspecified"}
	}
	return GatewayStatusDTO{
		State:                  ServiceState(status.GetState()),
		TargetAddr:             status.GetTargetAddr(),
		ChecksumMode:           ChecksumMode(status.GetChecksumMode()),
		PollingIntervalMs:      status.GetPollingIntervalMs(),
		RequestTimeoutMs:       status.GetRequestTimeoutMs(),
		ConnectTimeoutMs:       status.GetConnectTimeoutMs(),
		Connected:              status.GetConnected(),
		ConnectionAttempts:     int64(status.GetConnectionAttempts()),
		SuccessfulReads:        int64(status.GetSuccessfulReads()),
		FailedReads:            int64(status.GetFailedReads()),
		Reconnects:             int64(status.GetReconnects()),
		LastSuccessfulReadTime: Timestamp(status.GetLastSuccessfulReadTime()),
		LastDeviceTime:         Timestamp(status.GetLastDeviceTime()),
		LastError:              status.GetLastError(),
		LastTxTime:             Timestamp(status.GetLastTxTime()),
		LastRxTime:             Timestamp(status.GetLastRxTime()),
		RecentFramesCount:      int64(status.GetRecentFramesCount()),
	}
}

func LastReadTime(read *ft12v1.GetLastReadTimeResponse) LastReadTimeDTO {
	if read == nil {
		return LastReadTimeDTO{}
	}
	return LastReadTimeDTO{
		Available:  read.GetAvailable(),
		DeviceTime: Timestamp(read.GetDeviceTime()),
		ReadTime:   Timestamp(read.GetReadTime()),
	}
}
