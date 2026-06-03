package mapping

import ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"

func ServiceState(running bool, lastError string) ft12v1.ServiceState {
	state := ft12v1.ServiceState_SERVICE_STATE_STOPPED
	if running {
		state = ft12v1.ServiceState_SERVICE_STATE_RUNNING
	}
	if lastError != "" && running {
		state = ft12v1.ServiceState_SERVICE_STATE_DEGRADED
	}
	return state
}
