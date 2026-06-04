package dto

import (
	"strings"
	"time"

	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type HealthDTO struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Version   string `json:"version"`
	Commit    string `json:"commit,omitempty"`
	BuildDate string `json:"buildDate,omitempty"`
}

type PublicConfigDTO struct {
	EmulatorGRPC string `json:"emulatorGrpc"`
	GatewayGRPC  string `json:"gatewayGrpc"`
	PollingNote  string `json:"pollingNote"`
}

func Timestamp(ts *timestamppb.Timestamp) *string {
	if ts == nil {
		return nil
	}
	t := ts.AsTime().UTC().Format(time.RFC3339Nano)
	return &t
}

func ServiceState(state ft12v1.ServiceState) string {
	switch state {
	case ft12v1.ServiceState_SERVICE_STATE_STOPPED:
		return "stopped"
	case ft12v1.ServiceState_SERVICE_STATE_RUNNING:
		return "running"
	case ft12v1.ServiceState_SERVICE_STATE_DEGRADED:
		return "degraded"
	default:
		return "unspecified"
	}
}

func ChecksumMode(mode ft12v1.ChecksumMode) string {
	switch mode {
	case ft12v1.ChecksumMode_CHECKSUM_MODE_SUM:
		return "sum"
	case ft12v1.ChecksumMode_CHECKSUM_MODE_CRC16:
		return "crc16"
	default:
		return "unspecified"
	}
}

func Direction(direction ft12v1.EventDirection) string {
	switch direction {
	case ft12v1.EventDirection_EVENT_DIRECTION_RX:
		return "RX"
	case ft12v1.EventDirection_EVENT_DIRECTION_TX:
		return "TX"
	case ft12v1.EventDirection_EVENT_DIRECTION_ERROR:
		return "ERR"
	case ft12v1.EventDirection_EVENT_DIRECTION_SYSTEM:
		return "SYSTEM"
	default:
		return "SYSTEM"
	}
}

func EventDirectionFromString(direction string) ft12v1.EventDirection {
	switch strings.ToUpper(direction) {
	case "RX":
		return ft12v1.EventDirection_EVENT_DIRECTION_RX
	case "TX":
		return ft12v1.EventDirection_EVENT_DIRECTION_TX
	case "ERR", "ERROR":
		return ft12v1.EventDirection_EVENT_DIRECTION_ERROR
	default:
		return ft12v1.EventDirection_EVENT_DIRECTION_SYSTEM
	}
}
