package emulator

import (
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/observability/events"
)

type Status struct {
	Running             bool
	ListenAddress       string
	ChecksumMode        string
	ActiveConnections   int
	TotalConnections    uint64
	TotalRequests       uint64
	TotalResponses      uint64
	TotalProtocolErrors uint64
	LastRequestTime     time.Time
	LastResponseTime    time.Time
	LastError           string
	FaultMode           FaultMode
	RecentFramesCount   int
}

type Snapshot struct {
	Status Status
	Recent []events.FrameRecord
}
