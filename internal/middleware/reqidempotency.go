package middleware

import (
	"net/http"
	"sync"
	"time"
)

type idempotencyEntry struct {
	status  int
	headers http.Header
	body    []byte
	at      time.Time
}

// RequestIdempotency caches responses keyed by the value of keyHeader so that
// retried requests with the same key receive the original response.
type RequestIdempotency struct {
	mu      sync.Mutex
	store   map[string]*idempotencyEntry
	ttl     time.Duration
	keyHeader string
}

// NewRequestIdempotency returns an idempotency middleware.
// keyHeader is the request header that carries the idempotency key (e.g. "Idempotency-Key").
// ttl controls how long a cached response is retained.
func NewRequestIdempotency(keyHeader string, ttl time.Duration) func(http.Handler) http.Handler {
	if keyHeader == "" {
		keyHeader = "Idempotency-Key"
	}
	ri := &RequestIdempotency{
		store:     make(map[string]*idempotencyEntry),
		ttl:       ttl,
		keyHeader: keyHeader,
	}
	return ri.handler
}

func (ri *RequestIdempotency) handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get(ri.keyHeader)
		if key == "" {
			next.ServeHTTP(w, r)
			return
		}

		ri.mu.Lock()
		if entry, ok := ri.store[key]; ok && time.Since(entry.at) < ri.ttl {
			ri.mu.Unlock()
			for k, vals := range entry.headers {
				for _, v := range vals {
					w.Header().Add(k, v)
				}
			}
			w.Header().Set("X-Idempotency-Replayed", "true")
			w.WriteHeader(entry.status)
			w.Write(entry.body) //nolint:errcheck
			return
		}
		ri.mu.Unlock()

		rec := NewResponseRecorder(w)
		next.ServeHTTP(rec, r)

		ri.mu.Lock()
		ri.store[key] = &idempotencyEntry{
			status:  rec.Status(),
			headers: rec.Header().Clone(),
			body:    rec.Body(),
			at:      time.Now(),
		}
		ri.mu.Unlock()
	})
}
