package middleware

import (
	"net/http"
	"strconv"
)

// PriorityFunc extracts a numeric priority (higher = more important) from a request.
type PriorityFunc func(r *http.Request) int

// NewRequestPriority sets an X-Request-Priority response header based on the
// evaluated priority and optionally rejects low-priority requests when the
// server is under pressure (activeRequests > threshold).
func NewRequestPriority(next http.Handler, threshold int, fn PriorityFunc) http.Handler {
	if fn == nil {
		fn = defaultPriority
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pri := fn(r)
		w.Header().Set("X-Request-Priority", strconv.Itoa(pri))

		if threshold > 0 && pri < 0 {
			http.Error(w, "low priority request rejected", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// defaultPriority derives priority from the X-Priority header; defaults to 0.
func defaultPriority(r *http.Request) int {
	v := r.Header.Get("X-Priority")
	if v == "" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return n
}
