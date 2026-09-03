package dto

import ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"

// ScheduleDTO describes when the next poll is due and why.
type ScheduleDTO struct {
	Mode        string  `json:"mode"`
	IntervalMs  int64   `json:"intervalMs"`
	OffsetMs    int64   `json:"offsetMs"`
	Description string  `json:"description,omitempty"`
	NextPollAt  *string `json:"nextPollAt,omitempty"`
}

// RetryDTO covers the in-session retry budget and how much of it has been
// spent since the gateway started.
type RetryDTO struct {
	Attempts       int32  `json:"attempts"`
	DelayMs        int64  `json:"delayMs"`
	MaxDelayMs     int64  `json:"maxDelayMs"`
	TotalRetries   uint64 `json:"totalRetries"`
	ExhaustedPolls uint64 `json:"exhaustedPolls"`
}

// ClockDTO is the measured device clock. Skew is signed and in milliseconds:
// positive means the device reads ahead of the gateway.
type ClockDTO struct {
	State               string  `json:"state"`
	SkewMs              int64   `json:"skewMs"`
	MedianSkewMs        int64   `json:"medianSkewMs"`
	MinSkewMs           int64   `json:"minSkewMs"`
	MaxSkewMs           int64   `json:"maxSkewMs"`
	DriftPerDayMs       int64   `json:"driftPerDayMs"`
	DriftDetermined     bool    `json:"driftDetermined"`
	DriftFit            float64 `json:"driftFit"`
	Samples             int32   `json:"samples"`
	WarnThresholdMs     int64   `json:"warnThresholdMs"`
	CriticalThresholdMs int64   `json:"criticalThresholdMs"`
	RoundTripMs         int64   `json:"roundTripMs"`
	UpdatedAt           *string `json:"updatedAt,omitempty"`
	ObservedSamples     uint64  `json:"observedSamples"`
	RejectedSamples     uint64  `json:"rejectedSamples"`
}

// HealthDeviceDTO is reachability with hysteresis, plus the latency
// distribution the decision was made on.
type HealthDeviceDTO struct {
	State                string  `json:"state"`
	Since                *string `json:"since,omitempty"`
	Availability         float64 `json:"availability"`
	WindowSamples        int32   `json:"windowSamples"`
	ConsecutiveFailures  int32   `json:"consecutiveFailures"`
	ConsecutiveSuccesses int32   `json:"consecutiveSuccesses"`
	LatencyP50Ms         int64   `json:"latencyP50Ms"`
	LatencyP95Ms         int64   `json:"latencyP95Ms"`
	LatencyP99Ms         int64   `json:"latencyP99Ms"`
	LatencyMaxMs         int64   `json:"latencyMaxMs"`
	LatencyMeanMs        int64   `json:"latencyMeanMs"`
	DegradeAfter         int32   `json:"degradeAfter"`
	OfflineAfter         int32   `json:"offlineAfter"`
	RecoverAfter         int32   `json:"recoverAfter"`
}

// IdentityDTO is the device nameplate as the read-identity probe found it.
type IdentityDTO struct {
	Known     bool    `json:"known"`
	Supported bool    `json:"supported"`
	Model     string  `json:"model,omitempty"`
	Serial    string  `json:"serial,omitempty"`
	Firmware  string  `json:"firmware,omitempty"`
	ReadAt    *string `json:"readAt,omitempty"`
}

type GatewayStatusDTO struct {
	State                  string          `json:"state"`
	TargetAddr             string          `json:"targetAddr"`
	ChecksumMode           string          `json:"checksumMode"`
	PollingIntervalMs      int64           `json:"pollingIntervalMs"`
	RequestTimeoutMs       int64           `json:"requestTimeoutMs"`
	ConnectTimeoutMs       int64           `json:"connectTimeoutMs"`
	Connected              bool            `json:"connected"`
	ConnectionAttempts     int64           `json:"connectionAttempts"`
	SuccessfulReads        int64           `json:"successfulReads"`
	FailedReads            int64           `json:"failedReads"`
	Reconnects             int64           `json:"reconnects"`
	ProtocolErrors         int64           `json:"protocolErrors"`
	LastRoundTripMs        int64           `json:"lastRoundTripMs"`
	LastSuccessfulReadTime *string         `json:"lastSuccessfulReadTime,omitempty"`
	LastDeviceTime         *string         `json:"lastDeviceTime,omitempty"`
	LastError              string          `json:"lastError,omitempty"`
	LastTxTime             *string         `json:"lastTxTime,omitempty"`
	LastRxTime             *string         `json:"lastRxTime,omitempty"`
	RecentFramesCount      int64           `json:"recentFramesCount"`
	DeviceID               string          `json:"deviceId,omitempty"`
	DeviceName             string          `json:"deviceName,omitempty"`
	Schedule               ScheduleDTO     `json:"schedule"`
	Retry                  RetryDTO        `json:"retry"`
	Clock                  ClockDTO        `json:"clock"`
	Health                 HealthDeviceDTO `json:"health"`
	Identity               IdentityDTO     `json:"identity"`
}

// FleetSummaryDTO counts every polled device by state.
type FleetSummaryDTO struct {
	Devices            int32  `json:"devices"`
	Running            int32  `json:"running"`
	Online             int32  `json:"online"`
	Degraded           int32  `json:"degraded"`
	Offline            int32  `json:"offline"`
	Unknown            int32  `json:"unknown"`
	ClockOK            int32  `json:"clockOk"`
	ClockWarn          int32  `json:"clockWarn"`
	ClockCritical      int32  `json:"clockCritical"`
	ClockUnknown       int32  `json:"clockUnknown"`
	WorstClockSkewMs   int64  `json:"worstClockSkewMs"`
	WorstClockDeviceID string `json:"worstClockDeviceId,omitempty"`
}

type FleetDTO struct {
	Summary FleetSummaryDTO    `json:"summary"`
	Devices []GatewayStatusDTO `json:"devices"`
}

type LastReadTimeDTO struct {
	Available  bool    `json:"available"`
	DeviceTime *string `json:"deviceTime,omitempty"`
	ReadTime   *string `json:"readTime,omitempty"`
}

// unavailableGatewayStatus is what the API reports when the gateway answered
// with nothing. It is a valid, fully-shaped status rather than a null, so a
// consumer never has to test for absence before reading a field.
func unavailableGatewayStatus() GatewayStatusDTO {
	return GatewayStatusDTO{
		State:        "unavailable",
		ChecksumMode: "unspecified",
		Clock:        ClockDTO{State: "unknown"},
		Health:       HealthDeviceDTO{State: "unknown"},
	}
}

func GatewayStatus(status *ft12v1.GatewayStatus) GatewayStatusDTO {
	if status == nil {
		return unavailableGatewayStatus()
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
		ProtocolErrors:         int64(status.GetProtocolErrors()),
		LastRoundTripMs:        status.GetLastRoundTripMs(),
		LastSuccessfulReadTime: Timestamp(status.GetLastSuccessfulReadTime()),
		LastDeviceTime:         Timestamp(status.GetLastDeviceTime()),
		LastError:              status.GetLastError(),
		LastTxTime:             Timestamp(status.GetLastTxTime()),
		LastRxTime:             Timestamp(status.GetLastRxTime()),
		RecentFramesCount:      int64(status.GetRecentFramesCount()),
		DeviceID:               status.GetDeviceId(),
		DeviceName:             status.GetDeviceName(),
		Schedule:               schedule(status.GetSchedule()),
		Retry:                  retry(status.GetRetry()),
		Clock:                  clockStatus(status.GetClock()),
		Health:                 deviceHealth(status.GetHealth()),
		Identity:               identity(status.GetIdentity()),
	}
}

func schedule(in *ft12v1.ScheduleStatus) ScheduleDTO {
	return ScheduleDTO{
		Mode:        in.GetMode(),
		IntervalMs:  in.GetIntervalMs(),
		OffsetMs:    in.GetOffsetMs(),
		Description: in.GetDescription(),
		NextPollAt:  Timestamp(in.GetNextPollAt()),
	}
}

func retry(in *ft12v1.RetryStatus) RetryDTO {
	return RetryDTO{
		Attempts:       in.GetAttempts(),
		DelayMs:        in.GetDelayMs(),
		MaxDelayMs:     in.GetMaxDelayMs(),
		TotalRetries:   in.GetTotalRetries(),
		ExhaustedPolls: in.GetExhaustedPolls(),
	}
}

// clockStatus and deviceHealth default their state to "unknown" rather than
// the empty string: the console keys colour and wording off these values, and
// an empty string is not a state anything can name.
func clockStatus(in *ft12v1.ClockStatus) ClockDTO {
	out := ClockDTO{
		State:               in.GetState(),
		SkewMs:              in.GetSkewMs(),
		MedianSkewMs:        in.GetMedianSkewMs(),
		MinSkewMs:           in.GetMinSkewMs(),
		MaxSkewMs:           in.GetMaxSkewMs(),
		DriftPerDayMs:       in.GetDriftPerDayMs(),
		DriftDetermined:     in.GetDriftDetermined(),
		DriftFit:            in.GetDriftFit(),
		Samples:             in.GetSamples(),
		WarnThresholdMs:     in.GetWarnThresholdMs(),
		CriticalThresholdMs: in.GetCriticalThresholdMs(),
		RoundTripMs:         in.GetRoundTripMs(),
		UpdatedAt:           Timestamp(in.GetUpdatedAt()),
		ObservedSamples:     in.GetObservedSamples(),
		RejectedSamples:     in.GetRejectedSamples(),
	}
	if out.State == "" {
		out.State = "unknown"
	}
	return out
}

func deviceHealth(in *ft12v1.DeviceHealth) HealthDeviceDTO {
	out := HealthDeviceDTO{
		State:                in.GetState(),
		Since:                Timestamp(in.GetSince()),
		Availability:         in.GetAvailability(),
		WindowSamples:        in.GetWindowSamples(),
		ConsecutiveFailures:  in.GetConsecutiveFailures(),
		ConsecutiveSuccesses: in.GetConsecutiveSuccesses(),
		LatencyP50Ms:         in.GetLatencyP50Ms(),
		LatencyP95Ms:         in.GetLatencyP95Ms(),
		LatencyP99Ms:         in.GetLatencyP99Ms(),
		LatencyMaxMs:         in.GetLatencyMaxMs(),
		LatencyMeanMs:        in.GetLatencyMeanMs(),
		DegradeAfter:         in.GetDegradeAfter(),
		OfflineAfter:         in.GetOfflineAfter(),
		RecoverAfter:         in.GetRecoverAfter(),
	}
	if out.State == "" {
		out.State = "unknown"
	}
	return out
}

func identity(in *ft12v1.DeviceIdentity) IdentityDTO {
	return IdentityDTO{
		Known:     in.GetKnown(),
		Supported: in.GetSupported(),
		Model:     in.GetModel(),
		Serial:    in.GetSerial(),
		Firmware:  in.GetFirmware(),
		ReadAt:    Timestamp(in.GetReadAt()),
	}
}

func Fleet(fleet *ft12v1.GetFleetResponse) FleetDTO {
	out := FleetDTO{Devices: []GatewayStatusDTO{}}
	if fleet == nil {
		return out
	}
	summary := fleet.GetSummary()
	out.Summary = FleetSummaryDTO{
		Devices:            summary.GetDevices(),
		Running:            summary.GetRunning(),
		Online:             summary.GetOnline(),
		Degraded:           summary.GetDegraded(),
		Offline:            summary.GetOffline(),
		Unknown:            summary.GetUnknown(),
		ClockOK:            summary.GetClockOk(),
		ClockWarn:          summary.GetClockWarn(),
		ClockCritical:      summary.GetClockCritical(),
		ClockUnknown:       summary.GetClockUnknown(),
		WorstClockSkewMs:   summary.GetWorstClockSkewMs(),
		WorstClockDeviceID: summary.GetWorstClockDeviceId(),
	}
	for _, device := range fleet.GetDevices() {
		out.Devices = append(out.Devices, GatewayStatus(device))
	}
	return out
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
