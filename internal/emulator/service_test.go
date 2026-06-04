package emulator

import (
	"errors"
	"io"
	"log"
	"testing"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/observability/events"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
)

func testService(t *testing.T) *Service {
	t.Helper()
	cfg := config.DefaultEmulator()
	cfg.Listen = "127.0.0.1:0"
	cfg.ReadTimeoutDuration = time.Second
	cfg.WriteTimeoutDuration = time.Second
	service, err := NewService(&cfg, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	service.now = func() time.Time { return time.Unix(10, 0).UTC() }
	return service
}

func TestSetFaultModeClampsProbabilities(t *testing.T) {
	service := testService(t)

	got := service.SetFaultMode(FaultMode{BadChecksumProb: -1, FragmentProb: 2})
	if got.BadChecksumProb != 0 || got.FragmentProb != 1 {
		t.Fatalf("SetFaultMode() = %+v", got)
	}
	if service.Status().FaultMode.BadChecksumProb != 0 || service.Status().FaultMode.FragmentProb != 1 {
		t.Fatalf("status fault mode = %+v", service.Status().FaultMode)
	}
}

func TestHistoryRecordHelpersUpdateStatus(t *testing.T) {
	service := testService(t)
	req := frame.New(0x00, 0x01, []byte{0x01})
	raw, err := frame.Encode(req, checksum.ModeSum)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := frame.Decode(raw, checksum.ModeSum)
	if err != nil {
		t.Fatal(err)
	}

	service.recordRX("remote", decoded)
	service.recordTX("remote", raw, "read-time")
	service.recordError("remote", errors.New("bad frame"))

	status := service.Status()
	if status.TotalRequests != 1 || status.TotalResponses != 1 || status.TotalProtocolErrors != 1 {
		t.Fatalf("status counters = %+v", status)
	}
	if status.RecentFramesCount != 3 {
		t.Fatalf("recent count = %d, want 3", status.RecentFramesCount)
	}
	records := service.Snapshot().Recent
	if records[0].Direction != events.DirectionRX || records[1].Direction != events.DirectionTX || records[2].Direction != events.DirectionError {
		t.Fatalf("directions = %+v", records)
	}
}
