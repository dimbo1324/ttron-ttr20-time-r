package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	httpclient "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/http/client"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/api/http/dto"
	httperrors "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/http/errors"
)

type Config struct {
	RequestTimeout time.Duration
	EmulatorGRPC   string
	GatewayGRPC    string
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
	return mux
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	httperrors.WriteJSON(w, http.StatusOK, dto.HealthDTO{Status: "ok", Service: "ft12-api", Version: "dev"})
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

	emulatorStatus, err := h.emulator.GetStatus(ctx)
	if err != nil {
		httperrors.WriteUpstreamError(w, "EMULATOR", err)
		return
	}
	gatewayStatus, err := h.gateway.GetStatus(ctx)
	if err != nil {
		httperrors.WriteUpstreamError(w, "GATEWAY", err)
		return
	}
	lastRead, err := h.gateway.GetLastReadTime(ctx)
	if err != nil {
		httperrors.WriteUpstreamError(w, "GATEWAY", err)
		return
	}
	events, note := h.mergedEvents(ctx, 20)
	httperrors.WriteJSON(w, http.StatusOK, dto.OverviewDTO{
		Health:     dto.HealthDTO{Status: "ok", Service: "ft12-api", Version: "dev"},
		Emulator:   dto.EmulatorStatus(emulatorStatus),
		Gateway:    dto.GatewayStatus(gatewayStatus),
		LastRead:   dto.LastReadTime(lastRead),
		Events:     events,
		EventsNote: note,
	})
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
	events, err := h.emulator.GetRecentEvents(ctx, uint32(limit))
	if err != nil {
		httperrors.WriteUpstreamError(w, "EMULATOR", err)
		return
	}
	httperrors.WriteJSON(w, http.StatusOK, map[string]any{"events": dto.Events(events, "emulator")})
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
	events, err := h.gateway.GetRecentEvents(ctx, uint32(limit))
	if err != nil {
		httperrors.WriteUpstreamError(w, "GATEWAY", err)
		return
	}
	httperrors.WriteJSON(w, http.StatusOK, map[string]any{"events": dto.Events(events, "gateway")})
}

func (h *Handler) events(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	limit, ok := parseLimit(w, r, 100)
	if !ok {
		return
	}
	source := strings.ToLower(r.URL.Query().Get("source"))
	if source == "" {
		source = "all"
	}
	ctx, cancel := h.requestContext(r)
	defer cancel()

	var out []dto.EventDTO
	var note string
	switch source {
	case "all":
		out, note = h.mergedEvents(ctx, limit)
	case "emulator":
		events, err := h.emulator.GetRecentEvents(ctx, uint32(limit))
		if err != nil {
			httperrors.WriteUpstreamError(w, "EMULATOR", err)
			return
		}
		out = dto.Events(events, "emulator")
	case "gateway":
		events, err := h.gateway.GetRecentEvents(ctx, uint32(limit))
		if err != nil {
			httperrors.WriteUpstreamError(w, "GATEWAY", err)
			return
		}
		out = dto.Events(events, "gateway")
	default:
		httperrors.WriteError(w, http.StatusBadRequest, "INVALID_SOURCE", "source must be all, emulator, or gateway")
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

func (h *Handler) requestContext(r *http.Request) (context.Context, context.CancelFunc) {
	timeout := h.config.RequestTimeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return context.WithTimeout(r.Context(), timeout)
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
