package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

// retryAfterEntry tracks when a client may next be served.
type retryAfterEntry struct {
	until time.Time
}

// RetryAfterMiddleware enforces a per-IP back-off window after receiving a
// 429 response from the upstream. While the client is in back-off every
// subsequent request is rejected with 429 and a Retry-After header.
type RetryAfterMiddleware struct {
	mu      sync.Mutex
	clients map[string]*retryAfterEntry
	window  time.Duration
	now     func() time.Time
}

// NewRequestRetryAfter returns a middleware that enforces a back-off window
// (window) for clients that receive a 429 from upstream.
func NewRequestRetryAfter(window time.Duration) func(http.Handler) http.Handler {
	m := &RetryAfterMiddleware{
		clients: make(map[string]*retryAfterEntry),
		window:  window,
		now:     time.Now,
	}
	return m.Handler
}

func (m *RetryAfterMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realIP(r)

		m.mu.Lock()
		entry, ok := m.clients[ip]
		if ok && m.now().Before(entry.until) {
			retryIn := int(entry.until.Sub(m.now()).Seconds()) + 1
			m.mu.Unlock()
			w.Header().Set("Retry-After", strconv.Itoa(retryIn))
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		m.mu.Unlock()

		rec := NewResponseRecorder(w)
		next.ServeHTTP(rec, r)

		if rec.Status() == http.StatusTooManyRequests {
			m.mu.Lock()
			m.clients[ip] = &retryAfterEntry{until: m.now().Add(m.window)}
			m.mu.Unlock()
			w.Header().Set("Retry-After", strconv.Itoa(int(m.window.Seconds())))
		}
	})
}
