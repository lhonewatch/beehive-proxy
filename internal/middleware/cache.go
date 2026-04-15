package middleware

import (
	"net/http"
	"sync"
	"time"
)

// CacheEntry holds a cached response.
type CacheEntry struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	ExpiresAt  time.Time
}

// ResponseCache is a simple in-memory cache keyed by request path.
type ResponseCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	ttl     time.Duration
}

// NewResponseCache creates a new ResponseCache with the given TTL.
func NewResponseCache(ttl time.Duration) *ResponseCache {
	return &ResponseCache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
	}
}

// Get returns a cached entry and whether it was found and still valid.
func (c *ResponseCache) Get(key string) (*CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[key]
	if !ok || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}
	return entry, true
}

// Set stores a cache entry for the given key.
func (c *ResponseCache) Set(key string, entry *CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry.ExpiresAt = time.Now().Add(c.ttl)
	c.entries[key] = entry
}

// NewCacheMiddleware returns an HTTP middleware that caches GET responses.
func NewCacheMiddleware(cache *ResponseCache) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			key := r.URL.RequestURI()
			if entry, ok := cache.Get(key); ok {
				for k, vals := range entry.Headers {
					for _, v := range vals {
						w.Header().Add(k, v)
					}
				}
				w.Header().Set("X-Cache", "HIT")
				w.WriteHeader(entry.StatusCode)
				w.Write(entry.Body) //nolint:errcheck
				return
			}

			rec := NewResponseRecorder(w)
			next.ServeHTTP(rec, r)

			if rec.Status() < 500 {
				cache.Set(key, &CacheEntry{
					StatusCode: rec.Status(),
					Headers:    w.Header().Clone(),
					Body:       rec.Body(),
				})
			}
		})
	}
}
