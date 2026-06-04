package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	httperrors "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/http/errors"
)

type Logger interface {
	Printf(format string, v ...any)
}

type contextKey string

const requestIDKey contextKey = "request-id"

var requestCounter uint64

func Chain(handler http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	return handler
}

func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-ID")
			if id == "" {
				id = fmt.Sprintf("req-%d", atomic.AddUint64(&requestCounter, 1))
			}
			w.Header().Set("X-Request-ID", id)
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey, id)))
		})
	}
}

func SecurityHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Referrer-Policy", "no-referrer")
			next.ServeHTTP(w, r)
		})
	}
}

func Recovery(logger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					if logger != nil {
						logger.Printf("request panic id=%s method=%s path=%s panic=%v", RequestIDFromContext(r.Context()), r.Method, r.URL.Path, recovered)
					}
					httperrors.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "request failed")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func Logging(logger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			if logger != nil {
				logger.Printf("request id=%s method=%s path=%s status=%d duration=%s", RequestIDFromContext(r.Context()), r.Method, r.URL.Path, rec.status, time.Since(start))
			}
		})
	}
}

func Metrics(observe func(method, path string, status int, elapsed time.Duration)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			if observe != nil {
				observe(r.Method, r.URL.Path, rec.status, time.Since(start))
			}
		})
	}
}

func CORS(origin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type,X-Request-ID")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
