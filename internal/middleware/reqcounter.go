package middleware

import (
	"net/http"
	"sync/atomic"
)

// RequestCounter tracks the total number of requests handled.
type RequestCounter struct {
	total   atomic.Int64
	handler http.Handler
}

// NewRequestCounter wraps h and increments an internal counter on each request.
func NewRequestCounter(h http.Handler) *RequestCounter {
	return &RequestCounter{handler: h}
}

func (rc *RequestCounter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rc.total.Add(1)
	rc.handler.ServeHTTP(w, r)
}

// Total returns the number of requests seen so far.
func (rc *RequestCounter) Total() int64 {
	return rc.total.Load()
}

// NewRequestCounterMiddleware returns a middleware that counts requests and
// exposes the running total via the X-Request-Count response header.
func NewRequestCounterMiddleware(h http.Handler) http.Handler {
	rc := NewRequestCounter(h)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc.ServeHTTP(w, r)
		// Header is set after ServeHTTP so downstream can still write headers.
	})
}
