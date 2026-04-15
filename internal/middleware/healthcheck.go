package middleware

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

// HealthStatus represents the current health of the proxy.
type HealthStatus struct {
	Status    string `json:"status"`
	Uptime    string `json:"uptime"`
	Timestamp string `json:"timestamp"`
}

// HealthChecker tracks proxy liveness and exposes a /healthz endpoint.
type HealthChecker struct {
	startTime time.Time
	requests  atomic.Int64
}

// NewHealthChecker creates a new HealthChecker starting the uptime clock.
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{startTime: time.Now()}
}

// Handler returns an http.HandlerFunc that responds with JSON health status.
func (h *HealthChecker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := HealthStatus{
			Status:    "ok",
			Uptime:    time.Since(h.startTime).Round(time.Second).String(),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(status)
	}
}

// Middleware wraps the next handler, intercepting requests to /healthz.
func (h *HealthChecker) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			h.Handler()(w, r)
			return
		}
		h.requests.Add(1)
		next.ServeHTTP(w, r)
	})
}

// RequestCount returns the total number of proxied (non-health) requests.
func (h *HealthChecker) RequestCount() int64 {
	return h.requests.Load()
}
