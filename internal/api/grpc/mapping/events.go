package mapping

import (
	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/observability/events"
)

func Direction(direction events.Direction) ft12v1.EventDirection {
	switch direction {
	case events.DirectionRX:
		return ft12v1.EventDirection_EVENT_DIRECTION_RX
	case events.DirectionTX:
		return ft12v1.EventDirection_EVENT_DIRECTION_TX
	case events.DirectionError:
		return ft12v1.EventDirection_EVENT_DIRECTION_ERROR
	case events.DirectionSystem:
		return ft12v1.EventDirection_EVENT_DIRECTION_SYSTEM
	case events.DirectionDrop:
		return ft12v1.EventDirection_EVENT_DIRECTION_SYSTEM
	default:
		return ft12v1.EventDirection_EVENT_DIRECTION_SYSTEM
	}
}

func Events(records []events.FrameRecord, limit uint32) []*ft12v1.FrameEvent {
	if limit > 0 && int(limit) < len(records) {
		records = records[len(records)-int(limit):]
	}
	out := make([]*ft12v1.FrameEvent, 0, len(records))
	for _, record := range records {
		message := record.Message
		if message == "" {
			message = record.Error
		}
		out = append(out, &ft12v1.FrameEvent{
			Id:           record.ID,
			Timestamp:    Time(record.Timestamp),
			Service:      string(record.Service),
			Direction:    Direction(record.Direction),
			RemoteAddr:   record.RemoteAddr,
			ChecksumMode: ChecksumMode(record.ChecksumMode),
			RawHex:       record.RawHex,
			Command:      record.Command,
			Error:        record.Error,
			Message:      message,
		})
	}
	return out
}
