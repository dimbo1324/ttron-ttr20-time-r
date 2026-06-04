package gateway

import (
	"errors"
	"io"
	"log"
	"testing"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/observability/events"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/command"
)

func testGatewayService(t *testing.T) *Service {
	t.Helper()
	cfg := config.DefaultGateway()
	cfg.Target = "127.0.0.1:9000"
	service, err := NewService(&cfg, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	return service
}

func TestGatewayHistoryRecordHelpersUpdateStatus(t *testing.T) {
	service := testGatewayService(t)
	now := time.Unix(20, 0).UTC()

	service.recordTX("remote", []byte{0x01}, "read-time")
	service.recordRX("remote", []byte{0x02}, "read-time")
	service.recordProtocolError("remote", errors.New("bad checksum"))
	service.recordSuccess(command.ReadTimeResponse{Time: now})
	service.recordFailure(errors.New("timeout"))
	service.incrementConnectionAttempts()
	service.incrementReconnects()
	service.setConnected(true)
	service.setRunning(true)

	status := service.Status()
	if status.SuccessfulReads != 1 || status.FailedReads != 1 || status.ConnectionAttempts != 1 || status.ReconnectCount != 1 {
		t.Fatalf("status counters = %+v", status)
	}
	if !status.Connected || !status.Running {
		t.Fatalf("status flags = %+v", status)
	}
	if status.LastError != "timeout" {
		t.Fatalf("LastError = %q", status.LastError)
	}
	if status.RecentFramesCount != 3 {
		t.Fatalf("recent count = %d, want 3", status.RecentFramesCount)
	}
	records := service.Snapshot().Recent
	if records[0].Direction != events.DirectionTX || records[1].Direction != events.DirectionRX || records[2].Direction != events.DirectionError {
		t.Fatalf("directions = %+v", records)
	}
}
