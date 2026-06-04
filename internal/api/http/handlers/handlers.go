package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
	httpclient "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/http/client"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/api/http/dto"
	httperrors "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/http/errors"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/api/http/metrics"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/version"
)

type Config struct {
	RequestTimeout time.Duration
	EmulatorGRPC   string
	GatewayGRPC    string
	Metrics        *metrics.Registry
	MaxBodyBytes   int64
}

type Handler struct {
	emulator httpclient.EmulatorClient
	gateway  httpclient.GatewayClient
	config   Config
}

func New(emulator httpclient.EmulatorClient, gateway httpclient.GatewayClient, config Config) *Handler {
	return &Handler{emulator: emulator, gateway: gateway, config: config}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/api/v1/health", h.health)
	mux.HandleFunc("/api/v1/ready", h.ready)
	mux.HandleFunc("/metrics", h.metrics)
	mux.HandleFunc("/api/v1/config", h.configHandler)
	mux.HandleFunc("/api/v1/overview", h.overview)
	mux.HandleFunc("/api/v1/emulator/status", h.emulatorStatus)
	mux.HandleFunc("/api/v1/emulator/fault-mode", h.emulatorFaultMode)
	mux.HandleFunc("/api/v1/emulator/events", h.emulatorEvents)
	mux.HandleFunc("/api/v1/gateway/status", h.gatewayStatus)
	mux.HandleFunc("/api/v1/gateway/start", h.gatewayStart)
	mux.HandleFunc("/api/v1/gateway/stop", h.gatewayStop)
	mux.HandleFunc("/api/v1/gateway/last-read-time", h.gatewayLastReadTime)
	mux.HandleFunc("/api/v1/gateway/events", h.gatewayEvents)
	mux.HandleFunc("/api/v1/events", h.events)
	mux.HandleFunc("/api/v1/export/events.json", h.exportEventsJSON)
	mux.HandleFunc("/api/v1/export/events.csv", h.exportEventsCSV)
	mux.HandleFunc("/api/v1/export/overview.json", h.exportOverviewJSON)
	mux.HandleFunc("/api/v1/export/emulator-status.json", h.exportEmulatorStatusJSON)
	mux.HandleFunc("/api/v1/export/gateway-status.json", h.exportGatewayStatusJSON)
	return mux
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	httperrors.WriteJSON(w, http.StatusOK, healthDTO())
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	ctx, cancel := h.requestContext(r)
	defer cancel()

	emulatorState := "ok"
	gatewayState := "ok"
	status := http.StatusOK
	if _, err := h.emulator.GetStatus(ctx); err != nil {
		emulatorState = "unavailable"
		status = http.StatusServiceUnavailable
	}
	if _, err := h.gateway.GetStatus(ctx); err != nil {
		gatewayState = "unavailable"
		status = http.StatusServiceUnavailable
	}
	overall := "ready"
	if status != http.StatusOK {
		overall = "not_ready"
	}
	httperrors.WriteJSON(w, status, map[string]string{
		"status":   overall,
		"emulator": emulatorState,
		"gateway":  gatewayState,
	})
}

func (h *Handler) metrics(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	h.config.Metrics.Write(w)
}

func (h *Handler) configHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	httperrors.WriteJSON(w, http.StatusOK, dto.PublicConfigDTO{
		EmulatorGRPC: h.config.EmulatorGRPC,
		GatewayGRPC:  h.config.GatewayGRPC,
		PollingNote:  "UI polling uses HTTP/JSON over this API; the API uses internal gRPC clients.",
	})
}

func (h *Handler) overview(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	ctx, cancel := h.requestContext(r)
	defer cancel()

	overview, ok := h.buildOverview(w, ctx, 20)
	if !ok {
		return
	}
	httperrors.WriteJSON(w, http.StatusOK, overview)
}

func (h *Handler) emulatorStatus(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	ctx, cancel := h.requestContext(r)
	defer cancel()
	status, err := h.emulator.GetStatus(ctx)
	if err != nil {
		httperrors.WriteUpstreamError(w, "EMULATOR", err)
		return
	}
	httperrors.WriteJSON(w, http.StatusOK, dto.EmulatorStatus(status))
}

func (h *Handler) emulatorFaultMode(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getEmulatorFaultMode(w, r)
	case http.MethodPut:
		h.putEmulatorFaultMode(w, r)
	default:
		methodNotAllowed(w)
	}
}

func (h *Handler) getEmulatorFaultMode(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := h.requestContext(r)
	defer cancel()
	fault, err := h.emulator.GetFaultMode(ctx)
	if err != nil {
		httperrors.WriteUpstreamError(w, "EMULATOR", err)
		return
	}
	httperrors.WriteJSON(w, http.StatusOK, dto.FaultMode(fault))
}

func (h *Handler) putEmulatorFaultMode(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBodyBytes())
	var fault dto.FaultModeDTO
	if err := json.NewDecoder(r.Body).Decode(&fault); err != nil {
		httperrors.WriteError(w, http.StatusBadRequest, "BAD_JSON", "request body must be a valid fault mode JSON object")
		return
	}
	if err := validateFaultMode(fault); err != nil {
		httperrors.WriteError(w, http.StatusBadRequest, "INVALID_FAULT_MODE", err.Error())
		return
	}
	ctx, cancel := h.requestContext(r)
	defer cancel()
	next, status, err := h.emulator.SetFaultMode(ctx, dto.FaultModeProto(fault))
	if err != nil {
		httperrors.WriteUpstreamError(w, "EMULATOR", err)
		return
	}
	httperrors.WriteJSON(w, http.StatusOK, map[string]any{
		"faultMode": dto.FaultMode(next),
		"status":    dto.EmulatorStatus(status),
	})
}

func (h *Handler) emulatorEvents(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	limit, ok := parseLimit(w, r, 100)
	if !ok {
		return
	}
	ctx, cancel := h.requestContext(r)
	defer cancel()
	h.writeEvents(w, ctx, "emulator", "EMULATOR", h.emulator.GetRecentEvents, limit)
}

func (h *Handler) gatewayStatus(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	ctx, cancel := h.requestContext(r)
	defer cancel()
	status, err := h.gateway.GetStatus(ctx)
	if err != nil {
		httperrors.WriteUpstreamError(w, "GATEWAY", err)
		return
	}
	httperrors.WriteJSON(w, http.StatusOK, dto.GatewayStatus(status))
}

func (h *Handler) gatewayStart(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	ctx, cancel := h.requestContext(r)
	defer cancel()
	status, err := h.gateway.StartPolling(ctx)
	if err != nil {
		httperrors.WriteUpstreamError(w, "GATEWAY", err)
		return
	}
	httperrors.WriteJSON(w, http.StatusOK, map[string]any{"status": dto.GatewayStatus(status)})
}

func (h *Handler) gatewayStop(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	ctx, cancel := h.requestContext(r)
	defer cancel()
	status, err := h.gateway.StopPolling(ctx)
	if err != nil {
		httperrors.WriteUpstreamError(w, "GATEWAY", err)
		return
	}
	httperrors.WriteJSON(w, http.StatusOK, map[string]any{"status": dto.GatewayStatus(status)})
}

func (h *Handler) gatewayLastReadTime(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	ctx, cancel := h.requestContext(r)
	defer cancel()
	read, err := h.gateway.GetLastReadTime(ctx)
	if err != nil {
		httperrors.WriteUpstreamError(w, "GATEWAY", err)
		return
	}
	httperrors.WriteJSON(w, http.StatusOK, dto.LastReadTime(read))
}

func (h *Handler) gatewayEvents(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	limit, ok := parseLimit(w, r, 100)
	if !ok {
		return
	}
	ctx, cancel := h.requestContext(r)
	defer cancel()
	h.writeEvents(w, ctx, "gateway", "GATEWAY", h.gateway.GetRecentEvents, limit)
}

func (h *Handler) events(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	limit, ok := parseLimit(w, r, 100)
	if !ok {
		return
	}
	source := exportSource(r)
	ctx, cancel := h.requestContext(r)
	defer cancel()

	out, note, ok := h.eventsForSource(w, ctx, source, limit)
	if !ok {
		return
	}
	httperrors.WriteJSON(w, http.StatusOK, map[string]any{"events": out, "note": note})
}

func (h *Handler) mergedEvents(ctx context.Context, limit int) ([]dto.EventDTO, string) {
	emulatorEvents, emuErr := h.emulator.GetRecentEvents(ctx, uint32(limit))
	gatewayEvents, gwErr := h.gateway.GetRecentEvents(ctx, uint32(limit))
	note := ""
	if emuErr != nil {
		note = appendNote(note, "emulator events unavailable")
	}
	if gwErr != nil {
		note = appendNote(note, "gateway events unavailable")
	}
	return dto.MergeEvents(dto.Events(emulatorEvents, "emulator"), dto.Events(gatewayEvents, "gateway"), limit), note
}

func (h *Handler) writeEvents(w http.ResponseWriter, ctx context.Context, source, upstream string, fetch func(context.Context, uint32) ([]*ft12v1.FrameEvent, error), limit int) {
	events, err := fetch(ctx, uint32(limit))
	if err != nil {
		httperrors.WriteUpstreamError(w, upstream, err)
		return
	}
	httperrors.WriteJSON(w, http.StatusOK, map[string]any{"events": dto.Events(events, source)})
}

func (h *Handler) requestContext(r *http.Request) (context.Context, context.CancelFunc) {
	timeout := h.config.RequestTimeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return context.WithTimeout(r.Context(), timeout)
}

func (h *Handler) maxBodyBytes() int64 {
	if h.config.MaxBodyBytes > 0 {
		return h.config.MaxBodyBytes
	}
	return 64 * 1024
}

func parseLimit(w http.ResponseWriter, r *http.Request, fallback int) (int, bool) {
	raw := r.URL.Query().Get("limit")
	if raw == "" {
		return fallback, true
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit < 1 || limit > 1000 {
		httperrors.WriteError(w, http.StatusBadRequest, "INVALID_LIMIT", "limit must be an integer between 1 and 1000")
		return 0, false
	}
	return limit, true
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}
	methodNotAllowed(w)
	return false
}

func methodNotAllowed(w http.ResponseWriter) {
	httperrors.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "HTTP method is not supported for this endpoint")
}

func validateFaultMode(fault dto.FaultModeDTO) error {
	if fault.ResponseDelayMs < 0 {
		return fmt.Errorf("responseDelayMs must be non-negative")
	}
	if fault.FragmentDelayMs < 0 {
		return fmt.Errorf("fragmentDelayMs must be non-negative")
	}
	if fault.CorruptChecksumProbability < 0 || fault.CorruptChecksumProbability > 1 {
		return fmt.Errorf("corruptChecksumProbability must be in range 0..1")
	}
	if fault.FragmentProbability < 0 || fault.FragmentProbability > 1 {
		return fmt.Errorf("fragmentProbability must be in range 0..1")
	}
	return nil
}

func appendNote(current, next string) string {
	if current == "" {
		return next
	}
	return current + "; " + next
}

func healthDTO() dto.HealthDTO {
	return dto.HealthDTO{
		Status:    "ok",
		Service:   "ft12-api",
		Version:   version.Version,
		Commit:    version.Commit,
		BuildDate: version.BuildDate,
	}
}
