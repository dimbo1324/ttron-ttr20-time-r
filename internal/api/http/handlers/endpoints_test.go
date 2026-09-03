package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func doRequest(t *testing.T, handler *Handler, method, target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, nil)
	rec := httptest.NewRecorder()
	handler.Routes().ServeHTTP(rec, req)
	return rec
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not JSON: %v (body=%s)", err, rec.Body.String())
	}
	return body
}

func TestConfigEndpoint(t *testing.T) {
	handler, _, _ := testHandler()

	rec := doRequest(t, handler, http.MethodGet, "/api/v1/config")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	body := decodeBody(t, rec)
	if body["emulatorGrpc"] != "127.0.0.1:9100" || body["gatewayGrpc"] != "127.0.0.1:9200" {
		t.Fatalf("body = %v", body)
	}
}

func TestOverviewEndpoint(t *testing.T) {
	handler, _, _ := testHandler()

	rec := doRequest(t, handler, http.MethodGet, "/api/v1/overview")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	body := decodeBody(t, rec)
	if _, ok := body["emulator"]; !ok {
		t.Fatalf("overview must include the emulator: %v", body)
	}
	if _, ok := body["gateway"]; !ok {
		t.Fatalf("overview must include the gateway: %v", body)
	}
}

func TestOverviewEndpointReportsUpstreamFailure(t *testing.T) {
	handler, emulator, _ := testHandler()
	emulator.statusErr = errors.New("emulator unavailable")

	rec := doRequest(t, handler, http.MethodGet, "/api/v1/overview")
	if rec.Code == http.StatusOK {
		t.Fatalf("status = %d, want a failure", rec.Code)
	}
	if body := decodeBody(t, rec); body["error"] == nil {
		t.Fatalf("body = %s, want an error envelope", rec.Body.String())
	}
}

func TestEmulatorStatusEndpoint(t *testing.T) {
	handler, _, _ := testHandler()

	rec := doRequest(t, handler, http.MethodGet, "/api/v1/emulator/status")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if body := decodeBody(t, rec); body["listenAddr"] != "127.0.0.1:9000" {
		t.Fatalf("body = %v", body)
	}
}

func TestEmulatorStatusEndpointReportsUpstreamFailure(t *testing.T) {
	handler, emulator, _ := testHandler()
	emulator.statusErr = errors.New("emulator unavailable")

	rec := doRequest(t, handler, http.MethodGet, "/api/v1/emulator/status")
	if rec.Code == http.StatusOK {
		t.Fatalf("status = %d, want a failure", rec.Code)
	}
}

func TestGetEmulatorFaultMode(t *testing.T) {
	handler, _, _ := testHandler()

	rec := doRequest(t, handler, http.MethodGet, "/api/v1/emulator/fault-mode")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if _, ok := decodeBody(t, rec)["fragmentDelayMs"]; !ok {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestEmulatorFaultModeRejectsWrongMethod(t *testing.T) {
	handler, _, _ := testHandler()

	rec := doRequest(t, handler, http.MethodDelete, "/api/v1/emulator/fault-mode")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestEmulatorEventsEndpoint(t *testing.T) {
	handler, _, _ := testHandler()

	rec := doRequest(t, handler, http.MethodGet, "/api/v1/emulator/events?limit=10")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	body := decodeBody(t, rec)
	items, ok := body["events"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("body = %v", body)
	}
}

func TestEmulatorEventsRejectsInvalidLimit(t *testing.T) {
	handler, _, _ := testHandler()

	rec := doRequest(t, handler, http.MethodGet, "/api/v1/emulator/events?limit=abc")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestGatewayStatusEndpoint(t *testing.T) {
	handler, _, _ := testHandler()

	rec := doRequest(t, handler, http.MethodGet, "/api/v1/gateway/status")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if body := decodeBody(t, rec); body["targetAddr"] != "127.0.0.1:9000" {
		t.Fatalf("body = %v", body)
	}
}

func TestGatewayStatusEndpointReportsUpstreamFailure(t *testing.T) {
	handler, _, gateway := testHandler()
	gateway.statusErr = errors.New("gateway unavailable")

	rec := doRequest(t, handler, http.MethodGet, "/api/v1/gateway/status")
	if rec.Code == http.StatusOK {
		t.Fatalf("status = %d, want a failure", rec.Code)
	}
}

func TestGatewayLastReadTimeEndpoint(t *testing.T) {
	handler, _, _ := testHandler()

	rec := doRequest(t, handler, http.MethodGet, "/api/v1/gateway/last-read-time")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if body := decodeBody(t, rec); body["available"] != true {
		t.Fatalf("body = %v", body)
	}
}

func TestGatewayEventsEndpoint(t *testing.T) {
	handler, _, _ := testHandler()

	rec := doRequest(t, handler, http.MethodGet, "/api/v1/gateway/events")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	items, ok := decodeBody(t, rec)["events"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestGatewayFleetEndpoint(t *testing.T) {
	handler, _, _ := testHandler()

	rec := doRequest(t, handler, http.MethodGet, "/api/v1/gateway/fleet")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	body := decodeBody(t, rec)
	summary, ok := body["summary"].(map[string]any)
	if !ok || summary["devices"] != float64(1) {
		t.Fatalf("summary = %s", rec.Body.String())
	}
	devices, ok := body["devices"].([]any)
	if !ok || len(devices) != 1 {
		t.Fatalf("devices = %s", rec.Body.String())
	}
}

func TestGatewayFleetEndpointReportsUpstreamFailure(t *testing.T) {
	handler, _, gateway := testHandler()
	gateway.fleetErr = errors.New("gateway unavailable")

	rec := doRequest(t, handler, http.MethodGet, "/api/v1/gateway/fleet")
	if rec.Code == http.StatusOK {
		t.Fatalf("status = %d, want a failure", rec.Code)
	}
}

func TestGatewayFleetRejectsWrites(t *testing.T) {
	handler, _, _ := testHandler()

	rec := doRequest(t, handler, http.MethodPost, "/api/v1/gateway/fleet")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestGatewayEventsRejectsInvalidLimit(t *testing.T) {
	handler, _, _ := testHandler()

	rec := doRequest(t, handler, http.MethodGet, "/api/v1/gateway/events?limit=-5")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestEventsEndpointFiltersBySource(t *testing.T) {
	handler, _, _ := testHandler()

	tests := []struct {
		source string
		want   int
	}{
		{source: "emulator", want: 1},
		{source: "gateway", want: 1},
		{source: "all", want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			rec := doRequest(t, handler, http.MethodGet, "/api/v1/events?source="+tt.source)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
			}
			items, ok := decodeBody(t, rec)["events"].([]any)
			if !ok || len(items) != tt.want {
				t.Fatalf("events = %s, want %d entries", rec.Body.String(), tt.want)
			}
		})
	}
}

func TestExportEndpointsReportUpstreamFailure(t *testing.T) {
	targets := []string{
		"/api/v1/export/events.json",
		"/api/v1/export/events.csv",
		"/api/v1/export/overview.json",
		"/api/v1/export/emulator-status.json",
	}

	for _, target := range targets {
		t.Run(target, func(t *testing.T) {
			handler, emulator, _ := testHandler()
			emulator.statusErr = errors.New("emulator unavailable")

			rec := doRequest(t, handler, http.MethodGet, target)
			if rec.Code == http.StatusOK && target != "/api/v1/export/events.json" && target != "/api/v1/export/events.csv" {
				t.Fatalf("status = %d, want a failure for %s", rec.Code, target)
			}
		})
	}
}

func TestExportGatewayStatusReportsUpstreamFailure(t *testing.T) {
	handler, _, gateway := testHandler()
	gateway.statusErr = errors.New("gateway unavailable")

	rec := doRequest(t, handler, http.MethodGet, "/api/v1/export/gateway-status.json")
	if rec.Code == http.StatusOK {
		t.Fatalf("status = %d, want a failure", rec.Code)
	}
}

func TestExportEventsRejectsInvalidSource(t *testing.T) {
	handler, _, _ := testHandler()

	rec := doRequest(t, handler, http.MethodGet, "/api/v1/export/events.json?source=nope")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
