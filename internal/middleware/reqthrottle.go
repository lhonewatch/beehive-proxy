package middleware

import (
	"net/http"
	"sync"
	"time"
)

// tokenBucket holds state for a single IP.
type tokenBucket struct {
	tokens   float64
	lastSeen time.Time
}

// RequestThrottle delays requests that exceed a steady-state rate (tokens/sec)
// per remote IP using a token-bucket algorithm.
type RequestThrottle struct {
	rate     float64 // tokens replenished per second
	burst    float64 // maximum token capacity
	mu       sync.Mutex
	buckets  map[string]*tokenBucket
	next     http.Handler
}

// NewRequestThrottle returns a middleware that throttles requests per IP.
// rate is the sustained requests/sec allowed; burst is the maximum burst size.
func NewRequestThrottle(rate float64, burst float64, next http.Handler) http.Handler {
	return &RequestThrottle{
		rate:    rate,
		burst:   burst,
		buckets: make(map[string]*tokenBucket),
		next:    next,
	}
}

func (t *RequestThrottle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ip := realIP(r)
	now := time.Now()

	t.mu.Lock()
	b, ok := t.buckets[ip]
	if !ok {
		b = &tokenBucket{tokens: t.burst, lastSeen: now}
		t.buckets[ip] = b
	}
	elapsed := now.Sub(b.lastSeen).Seconds()
	b.tokens += elapsed * t.rate
	if b.tokens > t.burst {
		b.tokens = t.burst
	}
	b.lastSeen = now
	allowed := b.tokens >= 1.0
	if allowed {
		b.tokens -= 1.0
	}
	t.mu.Unlock()

	if !allowed {
		w.Header().Set("Retry-After", "1")
		http.Error(w, "too many requests", http.StatusTooManyRequests)
		return
	}
	t.next.ServeHTTP(w, r)
}
