package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/api/http/middleware"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type fakeEmulator struct {
	status *ft12v1.EmulatorStatus
	fault  *ft12v1.FaultMode
	events []*ft12v1.FrameEvent
}

func (f *fakeEmulator) GetStatus(context.Context) (*ft12v1.EmulatorStatus, error) {
	return f.status, nil
}

func (f *fakeEmulator) GetFaultMode(context.Context) (*ft12v1.FaultMode, error) {
	return f.fault, nil
}

func (f *fakeEmulator) SetFaultMode(_ context.Context, fault *ft12v1.FaultMode) (*ft12v1.FaultMode, *ft12v1.EmulatorStatus, error) {
	f.fault = fault
	return f.fault, f.status, nil
}

func (f *fakeEmulator) GetRecentEvents(context.Context, uint32) ([]*ft12v1.FrameEvent, error) {
	return f.events, nil
}

type fakeGateway struct {
	status  *ft12v1.GatewayStatus
	started bool
	stopped bool
	events  []*ft12v1.FrameEvent
}

func (f *fakeGateway) GetStatus(context.Context) (*ft12v1.GatewayStatus, error) {
	return f.status, nil
}

func (f *fakeGateway) StartPolling(context.Context) (*ft12v1.GatewayStatus, error) {
	f.started = true
	return f.status, nil
}

func (f *fakeGateway) StopPolling(context.Context) (*ft12v1.GatewayStatus, error) {
	f.stopped = true
	return f.status, nil
}

func (f *fakeGateway) GetLastReadTime(context.Context) (*ft12v1.GetLastReadTimeResponse, error) {
	now := timestamppb.Now()
	return &ft12v1.GetLastReadTimeResponse{Available: true, DeviceTime: now, ReadTime: now}, nil
}

func (f *fakeGateway) GetRecentEvents(context.Context, uint32) ([]*ft12v1.FrameEvent, error) {
	return f.events, nil
}

func testHandler() (*Handler, *fakeEmulator, *fakeGateway) {
	emulator := &fakeEmulator{
		status: &ft12v1.EmulatorStatus{State: ft12v1.ServiceState_SERVICE_STATE_RUNNING, ListenAddr: "127.0.0.1:9000", ChecksumMode: ft12v1.ChecksumMode_CHECKSUM_MODE_SUM},
		fault:  &ft12v1.FaultMode{FragmentDelayMs: 40},
		events: []*ft12v1.FrameEvent{{Id: 1, Timestamp: timestamppb.New(time.Unix(1, 0).UTC()), Service: "emulator", Direction: ft12v1.EventDirection_EVENT_DIRECTION_RX}},
	}
	gateway := &fakeGateway{
		status: &ft12v1.GatewayStatus{State: ft12v1.ServiceState_SERVICE_STATE_RUNNING, TargetAddr: "127.0.0.1:9000", ChecksumMode: ft12v1.ChecksumMode_CHECKSUM_MODE_SUM},
		events: []*ft12v1.FrameEvent{{Id: 2, Timestamp: timestamppb.New(time.Unix(2, 0).UTC()), Service: "gateway", Direction: ft12v1.EventDirection_EVENT_DIRECTION_TX}},
	}
	return New(emulator, gateway, Config{RequestTimeout: time.Second, EmulatorGRPC: "127.0.0.1:9100", GatewayGRPC: "127.0.0.1:9200"}), emulator, gateway
}

func TestHealthEndpoint(t *testing.T) {
	handler, _, _ := testHandler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"service":"ft12-api"`) {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestFaultModeValidation(t *testing.T) {
	handler, _, _ := testHandler()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/emulator/fault-mode", bytes.NewBufferString(`{"responseDelayMs":-1}`))
	rec := httptest.NewRecorder()
	handler.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"INVALID_FAULT_MODE"`) {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestGatewayStartStopHandlers(t *testing.T) {
	handler, _, gateway := testHandler()
	for _, path := range []string{"/api/v1/gateway/start", "/api/v1/gateway/stop"} {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		rec := httptest.NewRecorder()
		handler.Routes().ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d body=%s", path, rec.Code, rec.Body.String())
		}
	}
	if !gateway.started || !gateway.stopped {
		t.Fatalf("gateway started=%v stopped=%v", gateway.started, gateway.stopped)
	}
}

func TestEventsEndpointMergesEvents(t *testing.T) {
	handler, _, _ := testHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events?source=all&limit=10", nil)
	rec := httptest.NewRecorder()
	handler.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Events []struct {
			ID uint64 `json:"id"`
		} `json:"events"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Events) != 2 || body.Events[0].ID != 2 {
		t.Fatalf("events = %+v", body.Events)
	}
}

func TestCORSPreflight(t *testing.T) {
	handler, _, _ := testHandler()
	wrapped := middleware.CORS("http://localhost:5173")(handler.Routes())
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("cors origin = %q", got)
	}
}
