package gateway

import (
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/clock"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/health"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/observability/events"
)

type ScheduleStatus struct {
	Mode        string
	Interval    time.Duration
	Offset      time.Duration
	Description string
	NextPollAt  time.Time
}

type RetryStatus struct {
	Attempts       int
	Delay          time.Duration
	MaxDelay       time.Duration
	TotalRetries   uint64
	ExhaustedPolls uint64
}

type ClockStatus struct {
	State             string
	Skew              time.Duration
	MedianSkew        time.Duration
	MinSkew           time.Duration
	MaxSkew           time.Duration
	DriftPerDay       time.Duration
	DriftDetermined   bool
	DriftFit          float64
	Samples           int
	WarnThreshold     time.Duration
	CriticalThreshold time.Duration
	RoundTrip         time.Duration
	UpdatedAt         time.Time
	ObservedSamples   uint64
	RejectedSamples   uint64
}

type HealthStatus struct {
	State                string
	Since                time.Time
	Availability         float64
	WindowSamples        int
	ConsecutiveFailures  int
	ConsecutiveSuccesses int
	LatencyP50           time.Duration
	LatencyP95           time.Duration
	LatencyP99           time.Duration
	LatencyMax           time.Duration
	LatencyMean          time.Duration
	DegradeAfter         int
	OfflineAfter         int
	RecoverAfter         int
}

type IdentityStatus struct {
	Known     bool
	Supported bool
	Model     string
	Serial    string
	Firmware  string
	ReadAt    time.Time
}

type Status struct {
	Running                bool
	TargetAddress          string
	ChecksumMode           string
	PollingInterval        time.Duration
	RequestTimeout         time.Duration
	ConnectTimeout         time.Duration
	Connected              bool
	ConnectionAttempts     uint64
	SuccessfulReads        uint64
	FailedReads            uint64
	ReconnectCount         uint64
	ProtocolErrors         uint64
	LastSuccessfulReadTime time.Time
	LastParsedDeviceTime   time.Time
	LastError              string
	LastTXTimestamp        time.Time
	LastRXTimestamp        time.Time
	LastRoundTrip          time.Duration
	RecentFramesCount      int
	DeviceID               string
	DeviceName             string
	Schedule               ScheduleStatus
	Retry                  RetryStatus
	Clock                  ClockStatus
	Health                 HealthStatus
	Identity               IdentityStatus
}

type Snapshot struct {
	Status Status
	Recent []events.FrameRecord
}

// History is the contents of the two rolling windows, oldest first. It is kept
// out of Status deliberately: Status is copied on every read by callers that
// only want the current numbers, and carrying two slices along would make that
// read cost proportional to the window size.
type History struct {
	ClockSamples   []clock.Point
	HealthOutcomes []health.Outcome
}
