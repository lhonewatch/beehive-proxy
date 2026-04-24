package middleware

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	TraceIDHeader  = "X-Trace-ID"
	SpanIDHeader   = "X-Span-ID"
	TraceStartHeader = "X-Trace-Start"
)

// NewRequestTrace injects distributed tracing headers into every request.
// If X-Trace-ID is already present it is preserved; a new X-Span-ID is always
// generated so each hop gets a unique span. X-Trace-Start records the Unix
// nanosecond timestamp at which the trace began (set only when X-Trace-ID is
// absent).
func NewRequestTrace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := r.Clone(r.Context())

		// Preserve or generate a trace ID.
		if req.Header.Get(TraceIDHeader) == "" {
			req.Header.Set(TraceIDHeader, newTraceUUID())
			req.Header.Set(TraceStartHeader, formatNano(time.Now()))
		}

		// Always generate a fresh span ID for this hop.
		req.Header.Set(SpanIDHeader, newTraceUUID())

		// Propagate trace headers to the response so callers can correlate.
		w.Header().Set(TraceIDHeader, req.Header.Get(TraceIDHeader))
		w.Header().Set(SpanIDHeader, req.Header.Get(SpanIDHeader))

		next.ServeHTTP(w, req)
	})
}

func newTraceUUID() string {
	return uuid.NewString()
}

func formatNano(t time.Time) string {
	return strconv.FormatInt(t.UnixNano(), 10)
}
