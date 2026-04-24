package middleware

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// cacheEntry holds a cached response keyed by a request signature.
type cacheEntry struct {
	status  int
	headers http.Header
	body    []byte
	expiresAt time.Time
}

// RequestCache is a middleware that caches upstream responses per
// (method, path, query) key for a configurable TTL. Only GET requests
// with a 2xx response are cached.
type RequestCache struct {
	mu      sync.RWMutex
	store   map[string]*cacheEntry
	ttl     time.Duration
	now     func() time.Time
}

// NewRequestCache creates a RequestCache middleware with the given TTL.
func NewRequestCache(ttl time.Duration) *RequestCache {
	return &RequestCache{
		store: make(map[string]*cacheEntry),
		ttl:   ttl,
		now:   time.Now,
	}
}

// Handler returns an http.Handler that wraps next with caching logic.
func (rc *RequestCache) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		key := rc.cacheKey(r)

		rc.mu.RLock()
		entry, ok := rc.store[key]
		rc.mu.RUnlock()

		if ok && rc.now().Before(entry.expiresAt) {
			for k, vals := range entry.headers {
				for _, v := range vals {
					w.Header().Add(k, v)
				}
			}
			w.Header().Set("X-Request-Cache", "HIT")
			w.WriteHeader(entry.status)
			_, _ = w.Write(entry.body)
			return
		}

		rec := NewResponseRecorder(w)
		next.ServeHTTP(rec, r)

		if rec.Status() >= 200 && rec.Status() < 300 {
			hdrs := make(http.Header)
			for k, v := range rec.Header() {
				hdrs[k] = v
			}
			rc.mu.Lock()
			rc.store[key] = &cacheEntry{
				status:    rec.Status(),
				headers:   hdrs,
				body:      rec.Body(),
				expiresAt: rc.now().Add(rc.ttl),
			}
			rc.mu.Unlock()
		}
	})
}

func (rc *RequestCache) cacheKey(r *http.Request) string {
	raw := r.Method + "\x00" + r.URL.Path + "\x00" + r.URL.RawQuery
	sum := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", sum)
}
