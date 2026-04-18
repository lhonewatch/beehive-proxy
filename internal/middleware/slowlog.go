package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// NewSlowLog returns middleware that logs requests exceeding the given threshold.
func NewSlowLog(threshold time.Duration, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := NewResponseRecorder(w)
			next.ServeHTTP(rec, r)
			elapsed := time.Since(start)
			if elapsed >= threshold {
				logger.Warn("slow request",
					"method", r.Method,
					"path", r.URL.Path,
					"status", rec.Status(),
					"duration_ms", elapsed.Milliseconds(),
					"threshold_ms", threshold.Milliseconds(),
					"trace_id", r.Header.Get("X-Trace-ID"),
				)
			}
		})
	}
}
