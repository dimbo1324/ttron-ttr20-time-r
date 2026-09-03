package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func doRequest(t *testing.T, handler *Handler, method, target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, nil)
	rec := httptest.NewRecorder()
	handler.Routes().ServeHTTP(rec, req)
	return rec
}

// doRequestWithBody is the same, for the endpoints that take one. The body is
// written as the JSON the API actually receives rather than marshalled from a
// DTO, so a renamed field fails here rather than passing against itself.
func doRequestWithBody(t *testing.T, handler *Handler, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
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

func TestGatewaySettingsEndpoint(t *testing.T) {
	handler, _, gateway := testHandler()

	rec := doRequestWithBody(t, handler, http.MethodPut, "/api/v1/gateway/settings",
		`{"scheduleMode":"interval","pollIntervalMs":2000,"requestTimeoutMs":500,"retryAttempts":4,"retryDelayMs":100,"clockWarnMs":800,"clockCriticalMs":9000,"degradeAfter":2,"offlineAfter":6,"recoverAfter":1}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	if gateway.settings.GetPollIntervalMs() != 2000 || gateway.settings.GetRetryAttempts() != 4 {
		t.Fatalf("gateway received %+v", gateway.settings)
	}
	body := decodeBody(t, rec)
	// Both halves: what was applied, and the status it produced, so a console
	// redraws without a second round trip.
	if _, ok := body["settings"].(map[string]any); !ok {
		t.Fatalf("body = %s", rec.Body.String())
	}
	if _, ok := body["status"].(map[string]any); !ok {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestGatewaySettingsRejectsMalformedJSON(t *testing.T) {
	handler, _, _ := testHandler()

	rec := doRequestWithBody(t, handler, http.MethodPut, "/api/v1/gateway/settings", "{not json")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestGatewaySettingsReportsARejectedSettingAsABadRequest(t *testing.T) {
	handler, _, gateway := testHandler()
	gateway.settingsErr = status.Error(codes.InvalidArgument, "invalid gateway settings: request timeout must be below the poll interval")

	rec := doRequestWithBody(t, handler, http.MethodPut, "/api/v1/gateway/settings", `{"pollIntervalMs":100}`)

	// 400, not 502: the caller asked for something impossible, and reporting
	// it as an upstream failure would send the reader after a broken process.
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "below the poll interval") {
		t.Fatalf("body = %s, want the reason to survive", rec.Body.String())
	}
}

func TestGatewaySettingsReportsAnUnreachableGateway(t *testing.T) {
	handler, _, gateway := testHandler()
	gateway.settingsErr = status.Error(codes.Unavailable, "gateway is not running")

	rec := doRequestWithBody(t, handler, http.MethodPut, "/api/v1/gateway/settings", `{"pollIntervalMs":2000}`)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestGatewaySettingsRejectsReads(t *testing.T) {
	handler, _, _ := testHandler()

	rec := doRequest(t, handler, http.MethodGet, "/api/v1/gateway/settings")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestGatewayHistoryEndpoint(t *testing.T) {
	handler, _, _ := testHandler()

	rec := doRequest(t, handler, http.MethodGet, "/api/v1/gateway/history")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	body := decodeBody(t, rec)
	samples, ok := body["clockSamples"].([]any)
	if !ok || len(samples) != 1 {
		t.Fatalf("clockSamples = %s", rec.Body.String())
	}
	outcomes, ok := body["healthOutcomes"].([]any)
	if !ok || len(outcomes) != 1 {
		t.Fatalf("healthOutcomes = %s", rec.Body.String())
	}
}

func TestGatewayHistoryEndpointReportsUpstreamFailure(t *testing.T) {
	handler, _, gateway := testHandler()
	gateway.historyErr = errors.New("gateway unavailable")

	rec := doRequest(t, handler, http.MethodGet, "/api/v1/gateway/history")
	if rec.Code == http.StatusOK {
		t.Fatalf("status = %d, want a failure", rec.Code)
	}
}

func TestGatewayHistoryRejectsWrites(t *testing.T) {
	handler, _, _ := testHandler()

	rec := doRequest(t, handler, http.MethodPost, "/api/v1/gateway/history")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
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
