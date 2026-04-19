package middleware

import (
	"net/http"
	"strconv"
	"time"
)

// NewRequestLatency injects an X-Request-Latency header into the response
// containing the upstream round-trip duration in milliseconds. An optional
// threshold (>0) causes requests that exceed it to also receive an
// X-Request-Latency-Slow: true header.
func NewRequestLatency(threshold time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := NewResponseRecorder(w)
			next.ServeHTTP(rec, r)
			elapsed := time.Since(start)
			ms := elapsed.Milliseconds()
			w.Header().Set("X-Request-Latency", strconv.FormatInt(ms, 10)+"ms")
			if threshold > 0 && elapsed > threshold {
				w.Header().Set("X-Request-Latency-Slow", "true")
			}
			rec.WriteToWrapped()
		})
	}
}
