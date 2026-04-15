package tracing

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

const (
	// TraceIDHeader is the HTTP header used to propagate trace identifiers.
	TraceIDHeader = "X-Trace-Id"
)

// TracerFunc is a function that returns a trace ID for a given request.
// If the request already carries a trace ID it should be returned as-is.
type TracerFunc func(r *http.Request) string

// DefaultTracerFunc returns the existing trace ID from the incoming
// request, or generates a new one when none is present.
func DefaultTracerFunc(r *http.Request) string {
	if id := r.Header.Get(TraceIDHeader); id != "" {
		return id
	}
	return newTraceID()
}

// newTraceID generates a random hex trace identifier.
func newTraceID() string {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src) //nolint:gosec
	return fmt.Sprintf("%016x", r.Uint64())
}
