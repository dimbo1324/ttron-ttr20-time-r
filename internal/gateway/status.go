package gateway

import (
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/observability/events"
)

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
	LastSuccessfulReadTime time.Time
	LastParsedDeviceTime   time.Time
	LastError              string
	LastTXTimestamp        time.Time
	LastRXTimestamp        time.Time
	RecentFramesCount      int
}

type Snapshot struct {
	Status Status
	Recent []events.FrameRecord
}
