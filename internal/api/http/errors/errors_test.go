package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func decode(t *testing.T, rec *httptest.ResponseRecorder) Response {
	t.Helper()
	var body Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body %q: %v", rec.Body.String(), err)
	}
	return body
}

func TestWriteUpstreamErrorMapsAWrappedDeadline(t *testing.T) {
	rec := httptest.NewRecorder()

	// The shape a real call produces: a deadline that has passed through a
	// client and picked up context on the way. Comparing with == missed it,
	// and every timeout was reported as a bad gateway -- sending the reader
	// after a broken upstream instead of a slow one.
	WriteUpstreamError(rec, "GATEWAY", fmt.Errorf("gateway status: %w", context.DeadlineExceeded))

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusGatewayTimeout)
	}
	if got := decode(t, rec).Error.Code; got != "UPSTREAM_TIMEOUT" {
		t.Fatalf("code = %q", got)
	}
}

func TestWriteUpstreamErrorMapsABareDeadline(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteUpstreamError(rec, "GATEWAY", context.DeadlineExceeded)

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusGatewayTimeout)
	}
}

func TestWriteUpstreamErrorMapsGRPCCodes(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
		code string
	}{
		{
			name: "a process that is not running",
			err:  status.Error(codes.Unavailable, "connection refused"),
			want: http.StatusServiceUnavailable,
			code: "GATEWAY_UNAVAILABLE",
		},
		{
			name: "a call that ran out of time",
			err:  status.Error(codes.DeadlineExceeded, "too slow"),
			want: http.StatusGatewayTimeout,
			code: "UPSTREAM_TIMEOUT",
		},
		{
			// The caller asked for something the service will not do. Not a
			// failing upstream, and a 502 would send the reader hunting for a
			// broken process.
			name: "a request the service refused",
			err:  status.Error(codes.InvalidArgument, "request timeout must be below the poll interval"),
			want: http.StatusBadRequest,
			code: "INVALID_ARGUMENT",
		},
		{
			name: "anything else",
			err:  status.Error(codes.Internal, "boom"),
			want: http.StatusBadGateway,
			code: "GATEWAY_UPSTREAM_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			WriteUpstreamError(rec, "GATEWAY", tt.err)

			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
			if got := decode(t, rec).Error.Code; got != tt.code {
				t.Fatalf("code = %q, want %q", got, tt.code)
			}
		})
	}
}

func TestWriteUpstreamErrorKeepsTheServiceMessage(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteUpstreamError(rec, "GATEWAY", status.Error(codes.InvalidArgument, "offset must fit inside the interval"))

	// The reason has to survive the hop, or the console shows a status code
	// where it should show a sentence.
	if got := decode(t, rec).Error.Message; got != "offset must fit inside the interval" {
		t.Fatalf("message = %q", got)
	}
}

func TestWriteUpstreamErrorHandlesNothingAtAll(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteUpstreamError(rec, "GATEWAY", nil)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestWriteUpstreamErrorHandlesAPlainError(t *testing.T) {
	rec := httptest.NewRecorder()

	// Not every failure comes from gRPC; a bug in the adapter arrives as an
	// ordinary error and still has to be answered with valid JSON.
	WriteUpstreamError(rec, "EMULATOR", fmt.Errorf("adapter is broken"))

	if rec.Code == http.StatusOK {
		t.Fatalf("status = %d, want a failure", rec.Code)
	}
	if decode(t, rec).Error.Message == "" {
		t.Fatal("the message must survive")
	}
}

func TestWriteJSONSetsTheContentType(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteJSON(rec, http.StatusOK, map[string]string{"status": "ok"})

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q", got)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}
