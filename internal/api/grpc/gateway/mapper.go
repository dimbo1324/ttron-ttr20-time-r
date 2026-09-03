package gateway

import (
	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/mapping"
	domain "github.com/dimbo1324/ttron-ttr20-time-r/internal/gateway"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/observability/events"
)

func mapStatus(status domain.Status) *ft12v1.GatewayStatus {
	return &ft12v1.GatewayStatus{
		State:                  mapping.ServiceState(status.Running, status.LastError),
		TargetAddr:             status.TargetAddress,
		ChecksumMode:           mapping.ChecksumMode(status.ChecksumMode),
		PollingIntervalMs:      mapping.Millis(status.PollingInterval),
		RequestTimeoutMs:       mapping.Millis(status.RequestTimeout),
		ConnectTimeoutMs:       mapping.Millis(status.ConnectTimeout),
		Connected:              status.Connected,
		ConnectionAttempts:     status.ConnectionAttempts,
		SuccessfulReads:        status.SuccessfulReads,
		FailedReads:            status.FailedReads,
		Reconnects:             status.ReconnectCount,
		LastSuccessfulReadTime: mapping.Time(status.LastSuccessfulReadTime),
		LastDeviceTime:         mapping.Time(status.LastParsedDeviceTime),
		LastError:              status.LastError,
		LastTxTime:             mapping.Time(status.LastTXTimestamp),
		LastRxTime:             mapping.Time(status.LastRXTimestamp),
		RecentFramesCount:      uint32(status.RecentFramesCount),
		ProtocolErrors:         status.ProtocolErrors,
		LastRoundTripMs:        mapping.Millis(status.LastRoundTrip),
		DeviceId:               status.DeviceID,
		DeviceName:             status.DeviceName,
		Schedule:               mapSchedule(status.Schedule),
		Retry:                  mapRetry(status.Retry),
		Clock:                  mapClock(status.Clock),
		Health:                 mapHealth(status.Health),
		Identity:               mapIdentity(status.Identity),
	}
}

func mapSchedule(schedule domain.ScheduleStatus) *ft12v1.ScheduleStatus {
	return &ft12v1.ScheduleStatus{
		Mode:        schedule.Mode,
		IntervalMs:  mapping.Millis(schedule.Interval),
		OffsetMs:    mapping.Millis(schedule.Offset),
		Description: schedule.Description,
		NextPollAt:  mapping.Time(schedule.NextPollAt),
	}
}

func mapRetry(retry domain.RetryStatus) *ft12v1.RetryStatus {
	return &ft12v1.RetryStatus{
		Attempts:       int32(retry.Attempts),
		DelayMs:        mapping.Millis(retry.Delay),
		MaxDelayMs:     mapping.Millis(retry.MaxDelay),
		TotalRetries:   retry.TotalRetries,
		ExhaustedPolls: retry.ExhaustedPolls,
	}
}

func mapClock(clock domain.ClockStatus) *ft12v1.ClockStatus {
	return &ft12v1.ClockStatus{
		State:               clock.State,
		SkewMs:              mapping.Millis(clock.Skew),
		MedianSkewMs:        mapping.Millis(clock.MedianSkew),
		MinSkewMs:           mapping.Millis(clock.MinSkew),
		MaxSkewMs:           mapping.Millis(clock.MaxSkew),
		DriftPerDayMs:       mapping.Millis(clock.DriftPerDay),
		DriftDetermined:     clock.DriftDetermined,
		DriftFit:            clock.DriftFit,
		Samples:             int32(clock.Samples),
		WarnThresholdMs:     mapping.Millis(clock.WarnThreshold),
		CriticalThresholdMs: mapping.Millis(clock.CriticalThreshold),
		RoundTripMs:         mapping.Millis(clock.RoundTrip),
		UpdatedAt:           mapping.Time(clock.UpdatedAt),
		ObservedSamples:     clock.ObservedSamples,
		RejectedSamples:     clock.RejectedSamples,
	}
}

func mapHealth(health domain.HealthStatus) *ft12v1.DeviceHealth {
	return &ft12v1.DeviceHealth{
		State:                health.State,
		Since:                mapping.Time(health.Since),
		Availability:         health.Availability,
		WindowSamples:        int32(health.WindowSamples),
		ConsecutiveFailures:  int32(health.ConsecutiveFailures),
		ConsecutiveSuccesses: int32(health.ConsecutiveSuccesses),
		LatencyP50Ms:         mapping.Millis(health.LatencyP50),
		LatencyP95Ms:         mapping.Millis(health.LatencyP95),
		LatencyP99Ms:         mapping.Millis(health.LatencyP99),
		LatencyMaxMs:         mapping.Millis(health.LatencyMax),
		LatencyMeanMs:        mapping.Millis(health.LatencyMean),
		DegradeAfter:         int32(health.DegradeAfter),
		OfflineAfter:         int32(health.OfflineAfter),
		RecoverAfter:         int32(health.RecoverAfter),
	}
}

func mapIdentity(identity domain.IdentityStatus) *ft12v1.DeviceIdentity {
	return &ft12v1.DeviceIdentity{
		Known:     identity.Known,
		Supported: identity.Supported,
		Model:     identity.Model,
		Serial:    identity.Serial,
		Firmware:  identity.Firmware,
		ReadAt:    mapping.Time(identity.ReadAt),
	}
}

func mapFleet(fleet domain.FleetStatus) *ft12v1.GetFleetResponse {
	statuses := make([]*ft12v1.GatewayStatus, 0, len(fleet.Statuses))
	for _, status := range fleet.Statuses {
		statuses = append(statuses, mapStatus(status))
	}
	return &ft12v1.GetFleetResponse{
		Summary: &ft12v1.FleetSummary{
			Devices:            int32(fleet.Devices),
			Running:            int32(fleet.Running),
			Online:             int32(fleet.Online),
			Degraded:           int32(fleet.Degraded),
			Offline:            int32(fleet.Offline),
			Unknown:            int32(fleet.Unknown),
			ClockOk:            int32(fleet.ClockOK),
			ClockWarn:          int32(fleet.ClockWarn),
			ClockCritical:      int32(fleet.ClockCritical),
			ClockUnknown:       int32(fleet.ClockUnknown),
			WorstClockSkewMs:   mapping.Millis(fleet.WorstClockSkew),
			WorstClockDeviceId: fleet.WorstClockID,
		},
		Devices: statuses,
	}
}

func mapHistory(history domain.History) *ft12v1.GetHistoryResponse {
	samples := make([]*ft12v1.ClockSample, 0, len(history.ClockSamples))
	for _, sample := range history.ClockSamples {
		samples = append(samples, &ft12v1.ClockSample{
			At:          mapping.Time(sample.At),
			SkewMs:      mapping.Millis(sample.Skew),
			RoundTripMs: mapping.Millis(sample.RoundTrip),
		})
	}
	outcomes := make([]*ft12v1.HealthOutcome, 0, len(history.HealthOutcomes))
	for _, outcome := range history.HealthOutcomes {
		outcomes = append(outcomes, &ft12v1.HealthOutcome{
			At:        mapping.Time(outcome.At),
			Success:   outcome.Success,
			LatencyMs: mapping.Millis(outcome.Latency),
		})
	}
	return &ft12v1.GetHistoryResponse{ClockSamples: samples, HealthOutcomes: outcomes}
}

func mapEvents(records []events.FrameRecord, limit uint32) []*ft12v1.FrameEvent {
	return mapping.Events(records, limit)
}
