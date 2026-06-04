package handlers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/api/http/dto"
	httperrors "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/http/errors"
)

type eventsExportDTO struct {
	ExportedAt string         `json:"exportedAt"`
	Source     string         `json:"source"`
	Limit      int            `json:"limit"`
	Events     []dto.EventDTO `json:"events"`
	Note       string         `json:"note,omitempty"`
}

type overviewExportDTO struct {
	ExportedAt string          `json:"exportedAt"`
	Overview   dto.OverviewDTO `json:"overview"`
}

type statusExportDTO[T any] struct {
	ExportedAt string `json:"exportedAt"`
	Status     T      `json:"status"`
}

func (h *Handler) exportEventsJSON(w http.ResponseWriter, r *http.Request) {
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

	events, note, ok := h.eventsForSource(w, ctx, source, limit)
	if !ok {
		return
	}
	writeDownloadJSON(w, "ft12-events", eventsExportDTO{
		ExportedAt: exportTimestamp(),
		Source:     source,
		Limit:      limit,
		Events:     events,
		Note:       note,
	})
}

func (h *Handler) exportEventsCSV(w http.ResponseWriter, r *http.Request) {
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

	events, _, ok := h.eventsForSource(w, ctx, source, limit)
	if !ok {
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", downloadDisposition("ft12-events", "csv"))
	w.WriteHeader(http.StatusOK)
	_ = writeEventsCSV(w, events)
}

func (h *Handler) exportOverviewJSON(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	limit, ok := parseLimit(w, r, 50)
	if !ok {
		return
	}
	ctx, cancel := h.requestContext(r)
	defer cancel()

	overview, ok := h.buildOverview(w, ctx, limit)
	if !ok {
		return
	}
	writeDownloadJSON(w, "ft12-overview", overviewExportDTO{
		ExportedAt: exportTimestamp(),
		Overview:   overview,
	})
}

func (h *Handler) exportEmulatorStatusJSON(w http.ResponseWriter, r *http.Request) {
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
	writeDownloadJSON(w, "ft12-emulator-status", statusExportDTO[dto.EmulatorStatusDTO]{
		ExportedAt: exportTimestamp(),
		Status:     dto.EmulatorStatus(status),
	})
}

func (h *Handler) exportGatewayStatusJSON(w http.ResponseWriter, r *http.Request) {
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
	writeDownloadJSON(w, "ft12-gateway-status", statusExportDTO[dto.GatewayStatusDTO]{
		ExportedAt: exportTimestamp(),
		Status:     dto.GatewayStatus(status),
	})
}

func (h *Handler) buildOverview(w http.ResponseWriter, ctx context.Context, eventLimit int) (dto.OverviewDTO, bool) {
	emulatorStatus, err := h.emulator.GetStatus(ctx)
	if err != nil {
		httperrors.WriteUpstreamError(w, "EMULATOR", err)
		return dto.OverviewDTO{}, false
	}
	gatewayStatus, err := h.gateway.GetStatus(ctx)
	if err != nil {
		httperrors.WriteUpstreamError(w, "GATEWAY", err)
		return dto.OverviewDTO{}, false
	}
	lastRead, err := h.gateway.GetLastReadTime(ctx)
	if err != nil {
		httperrors.WriteUpstreamError(w, "GATEWAY", err)
		return dto.OverviewDTO{}, false
	}
	events, note := h.mergedEvents(ctx, eventLimit)
	return dto.OverviewDTO{
		Health:     healthDTO(),
		Emulator:   dto.EmulatorStatus(emulatorStatus),
		Gateway:    dto.GatewayStatus(gatewayStatus),
		LastRead:   dto.LastReadTime(lastRead),
		Events:     events,
		EventsNote: note,
	}, true
}

func (h *Handler) eventsForSource(w http.ResponseWriter, ctx context.Context, source string, limit int) ([]dto.EventDTO, string, bool) {
	switch source {
	case "all":
		events, note := h.mergedEvents(ctx, limit)
		return events, note, true
	case "emulator":
		events, err := h.emulator.GetRecentEvents(ctx, uint32(limit))
		if err != nil {
			httperrors.WriteUpstreamError(w, "EMULATOR", err)
			return nil, "", false
		}
		return dto.Events(events, "emulator"), "", true
	case "gateway":
		events, err := h.gateway.GetRecentEvents(ctx, uint32(limit))
		if err != nil {
			httperrors.WriteUpstreamError(w, "GATEWAY", err)
			return nil, "", false
		}
		return dto.Events(events, "gateway"), "", true
	default:
		httperrors.WriteError(w, http.StatusBadRequest, "INVALID_SOURCE", "source must be all, emulator, or gateway")
		return nil, "", false
	}
}

func writeEventsCSV(w io.Writer, events []dto.EventDTO) error {
	writer := csv.NewWriter(w)
	if err := writer.Write([]string{
		"timestamp",
		"source",
		"service",
		"direction",
		"command",
		"checksumMode",
		"remoteAddr",
		"rawHex",
		"message",
		"error",
	}); err != nil {
		return err
	}
	for _, event := range events {
		timestamp := ""
		if event.Timestamp != nil {
			timestamp = *event.Timestamp
		}
		if err := writer.Write([]string{
			timestamp,
			event.Source,
			event.Service,
			event.Direction,
			event.Command,
			event.ChecksumMode,
			event.RemoteAddr,
			event.RawHex,
			event.Message,
			event.Error,
		}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func writeDownloadJSON(w http.ResponseWriter, prefix string, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", downloadDisposition(prefix, "json"))
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(value)
}

func downloadDisposition(prefix, ext string) string {
	return `attachment; filename="` + prefix + "-" + time.Now().UTC().Format("20060102-150405") + "." + ext + `"`
}

func exportTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

func exportSource(r *http.Request) string {
	source := strings.ToLower(r.URL.Query().Get("source"))
	if source == "" {
		return "all"
	}
	return source
}
