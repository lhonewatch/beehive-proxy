package middleware

import (
	"net/http"
	"sync/atomic"
	"time"
)

// RequestBreather adds an artificial delay ("breathing room") to incoming
// requests based on current load. When in-flight requests exceed the soft
// limit the delay grows linearly up to maxDelay.
type RequestBreather struct {
	soft     int64
	maxDelay time.Duration
	inflight atomic.Int64
}

func NewRequestBreather(softLimit int, maxDelay time.Duration) *RequestBreather {
	return &RequestBreather{
		soft:     int64(softLimit),
		maxDelay: maxDelay,
	}
}

func (rb *RequestBreather) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := rb.inflight.Add(1)
		defer rb.inflight.Add(-1)

		if rb.soft > 0 && current > rb.soft {
			excess := current - rb.soft
			delay := time.Duration(excess) * rb.maxDelay / time.Duration(rb.soft)
			if delay > rb.maxDelay {
				delay = rb.maxDelay
			}
			select {
			case <-time.After(delay):
			case <-r.Context().Done():
				http.Error(w, "request cancelled", http.StatusServiceUnavailable)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
