package errors

import (
	"context"
	"encoding/json"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Response struct {
	Error Error `json:"error"`
}

type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func WriteJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(value)
}

func WriteError(w http.ResponseWriter, statusCode int, code, message string) {
	WriteJSON(w, statusCode, Response{Error: Error{Code: code, Message: message}})
}

func WriteUpstreamError(w http.ResponseWriter, service string, err error) {
	if err == nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "unexpected empty upstream error")
		return
	}
	if err == context.DeadlineExceeded {
		WriteError(w, http.StatusGatewayTimeout, "UPSTREAM_TIMEOUT", service+" request timed out")
		return
	}
	st, ok := status.FromError(err)
	if !ok {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	switch st.Code() {
	case codes.Unavailable:
		WriteError(w, http.StatusServiceUnavailable, service+"_UNAVAILABLE", service+" gRPC service is unavailable")
	case codes.DeadlineExceeded:
		WriteError(w, http.StatusGatewayTimeout, "UPSTREAM_TIMEOUT", service+" request timed out")
	default:
		WriteError(w, http.StatusBadGateway, service+"_UPSTREAM_ERROR", st.Message())
	}
}
