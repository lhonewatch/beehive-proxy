package middleware

import (
	"net/http"
	"sync"
	"time"
)

type inflightEntry struct {
	ch  chan struct{}
	code int
	header http.Header
	body []byte
}

// NewDedupe returns middleware that collapses concurrent identical GET requests
// into a single upstream call. Subsequent in-flight requests wait and share the
// first response. Keyed on method + URL.
func NewDedupe() func(http.Handler) http.Handler {
	var (
		mu      sync.Mutex
		inflight = make(map[string]*inflightEntry)
	)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			key := r.URL.String()

			mu.Lock()
			if e, ok := inflight[key]; ok {
				mu.Unlock()
				select {
				case <-e.ch:
				case <-time.After(10 * time.Second):
					http.Error(w, "dedupe timeout", http.StatusGatewayTimeout)
					return
				}
				for k, vals := range e.header {
					for _, v := range vals {
						w.Header().Add(k, v)
					}
				}
				w.Header().Set("X-Dedupe", "HIT")
				w.WriteHeader(e.code)
				w.Write(e.body) //nolint:errcheck
				return
			}

			e := &inflightEntry{ch: make(chan struct{})}
			inflight[key] = e
			mu.Unlock()

			rec := NewResponseRecorder(w)
			next.ServeHTTP(rec, r)

			e.code = rec.Status()
			e.body = rec.Body()
			e.header = rec.Header().Clone()
			close(e.ch)

			mu.Lock()
			delete(inflight, key)
			mu.Unlock()
		})
	}
}
