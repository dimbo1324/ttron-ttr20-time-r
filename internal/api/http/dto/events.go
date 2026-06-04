package dto

import (
	"sort"

	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
)

type EventDTO struct {
	ID           uint64  `json:"id"`
	Timestamp    *string `json:"timestamp,omitempty"`
	Source       string  `json:"source"`
	Service      string  `json:"service"`
	Direction    string  `json:"direction"`
	RemoteAddr   string  `json:"remoteAddr,omitempty"`
	ChecksumMode string  `json:"checksumMode"`
	RawHex       string  `json:"rawHex,omitempty"`
	Command      string  `json:"command,omitempty"`
	Error        string  `json:"error,omitempty"`
	Message      string  `json:"message,omitempty"`
}

type OverviewDTO struct {
	Health     HealthDTO         `json:"health"`
	Emulator   EmulatorStatusDTO `json:"emulator"`
	Gateway    GatewayStatusDTO  `json:"gateway"`
	LastRead   LastReadTimeDTO   `json:"lastRead"`
	Events     []EventDTO        `json:"events"`
	EventsNote string            `json:"eventsNote,omitempty"`
}

func Events(events []*ft12v1.FrameEvent, fallbackSource string) []EventDTO {
	out := make([]EventDTO, 0, len(events))
	for _, event := range events {
		out = append(out, Event(event, fallbackSource))
	}
	return out
}

func Event(event *ft12v1.FrameEvent, fallbackSource string) EventDTO {
	if event == nil {
		return EventDTO{Source: fallbackSource, Service: fallbackSource, Direction: "SYSTEM", ChecksumMode: "unspecified"}
	}
	source := event.GetService()
	if source == "" {
		source = fallbackSource
	}
	return EventDTO{
		ID:           event.GetId(),
		Timestamp:    Timestamp(event.GetTimestamp()),
		Source:       source,
		Service:      source,
		Direction:    Direction(event.GetDirection()),
		RemoteAddr:   event.GetRemoteAddr(),
		ChecksumMode: ChecksumMode(event.GetChecksumMode()),
		RawHex:       event.GetRawHex(),
		Command:      event.GetCommand(),
		Error:        event.GetError(),
		Message:      event.GetMessage(),
	}
}

func MergeEvents(emulatorEvents, gatewayEvents []EventDTO, limit int) []EventDTO {
	merged := make([]EventDTO, 0, len(emulatorEvents)+len(gatewayEvents))
	merged = append(merged, emulatorEvents...)
	merged = append(merged, gatewayEvents...)
	sort.SliceStable(merged, func(i, j int) bool {
		left := ""
		right := ""
		if merged[i].Timestamp != nil {
			left = *merged[i].Timestamp
		}
		if merged[j].Timestamp != nil {
			right = *merged[j].Timestamp
		}
		if left == right {
			return merged[i].ID > merged[j].ID
		}
		return left > right
	})
	if limit > 0 && len(merged) > limit {
		return merged[:limit]
	}
	return merged
}
