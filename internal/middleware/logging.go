package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// ResponseRecorder wraps http.ResponseWriter to capture the status code.
type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
}

func (r *ResponseRecorder) WriteHeader(code int) {
	r.StatusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// NewResponseRecorder returns a ResponseRecorder with a default 200 status.
func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{ResponseWriter: w, StatusCode: http.StatusOK}
}

// RequestLogger returns middleware that logs each request with method, path,
// status code, duration, and trace ID.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := NewResponseRecorder(w)

			next.ServeHTTP(rec, r)

			duration := time.Since(start)
			traceID := r.Header.Get("X-Trace-ID")

			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.StatusCode,
				"duration_ms", duration.Milliseconds(),
				"trace_id", traceID,
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}
