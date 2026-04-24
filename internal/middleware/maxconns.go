package middleware

import (
	"net/http"
	"sync/atomic"
)

// maxConns limits the number of concurrent requests. Requests exceeding the
// limit receive a 503 Service Unavailable response.
type maxConns struct {
	handler  http.Handler
	sem      chan struct{}
	rejected atomic.Int64
}

// NewMaxConns wraps h and allows at most n concurrent requests.
// If n <= 0 the handler is returned unwrapped.
func NewMaxConns(n int, h http.Handler) http.Handler {
	if n <= 0 {
		return h
	}
	return &maxConns{
		handler: h,
		sem:     make(chan struct{}, n),
	}
}

// Rejected returns the total number of requests that were rejected due to the
// concurrency limit being reached.
func (m *maxConns) Rejected() int64 {
	return m.rejected.Load()
}

func (m *maxConns) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	select {
	case m.sem <- struct{}{}:
		defer func() { <-m.sem }()
		m.handler.ServeHTTP(w, r)
	default:
		m.rejected.Add(1)
		http.Error(w, "too many concurrent requests", http.StatusServiceUnavailable)
	}
}
