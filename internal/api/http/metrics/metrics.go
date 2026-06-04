package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type Registry struct {
	mu       sync.RWMutex
	requests map[string]uint64
	duration map[string]float64
}

func NewRegistry() *Registry {
	return &Registry{
		requests: make(map[string]uint64),
		duration: make(map[string]float64),
	}
}

func (r *Registry) Observe(method, path string, status int, elapsed time.Duration) {
	if r == nil {
		return
	}
	key := metricKey(method, path, status)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests[key]++
	r.duration[key] += elapsed.Seconds()
}

func (r *Registry) Write(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	if r == nil {
		return
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	fmt.Fprintln(w, "# HELP ft12_http_requests_total Total HTTP requests handled by ft12-api.")
	fmt.Fprintln(w, "# TYPE ft12_http_requests_total counter")
	writeCounter(w, "ft12_http_requests_total", r.requests)
	fmt.Fprintln(w, "# HELP ft12_http_request_duration_seconds_total Total HTTP request duration in seconds.")
	fmt.Fprintln(w, "# TYPE ft12_http_request_duration_seconds_total counter")
	writeCounterFloat(w, "ft12_http_request_duration_seconds_total", r.duration)
}

func writeCounter(w http.ResponseWriter, name string, values map[string]uint64) {
	keys := sortedKeys(values)
	for _, key := range keys {
		method, path, status := splitMetricKey(key)
		fmt.Fprintf(w, "%s{method=%q,path=%q,status=%q} %d\n", name, method, path, status, values[key])
	}
}

func writeCounterFloat(w http.ResponseWriter, name string, values map[string]float64) {
	keys := sortedKeys(values)
	for _, key := range keys {
		method, path, status := splitMetricKey(key)
		fmt.Fprintf(w, "%s{method=%q,path=%q,status=%q} %.6f\n", name, method, path, status, values[key])
	}
}

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func metricKey(method, path string, status int) string {
	if path == "" {
		path = "/"
	}
	return fmt.Sprintf("%s\t%s\t%d", method, path, status)
}

func splitMetricKey(key string) (string, string, string) {
	parts := strings.Split(key, "\t")
	if len(parts) != 3 {
		return "unknown", "unknown", "0"
	}
	return parts[0], parts[1], parts[2]
}
