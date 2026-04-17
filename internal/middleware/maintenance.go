package middleware

import (
	"net/http"
	"sync/atomic"
)

// MaintenanceMode is a middleware that returns 503 when maintenance mode is enabled.
type MaintenanceMode struct {
	enabled atomic.Bool
	body    string
	next    http.Handler
}

// NewMaintenanceMode creates a new MaintenanceMode middleware.
func NewMaintenanceMode(body string) *MaintenanceMode {
	if body == "" {
		body = "Service temporarily unavailable. Please try again later."
	}
	return &MaintenanceMode{body: body}
}

// Enable turns on maintenance mode.
func (m *MaintenanceMode) Enable() {
	m.enabled.Store(true)
}

// Disable turns off maintenance mode.
func (m *MaintenanceMode) Disable() {
	m.enabled.Store(false)
}

// IsEnabled reports whether maintenance mode is currently active.
func (m *MaintenanceMode) IsEnabled() bool {
	return m.enabled.Load()
}

// Handler returns an http.Handler that wraps next with maintenance checking.
func (m *MaintenanceMode) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.enabled.Load() {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Header().Set("Retry-After", "120")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(m.body))
			return
		}
		next.ServeHTTP(w, r)
	})
}
