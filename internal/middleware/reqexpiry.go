package middleware

import (
	"net/http"
	"strconv"
	"time"
)

// NewRequestExpiry rejects requests whose X-Request-Timestamp header indicates
// they are older than maxAge. This guards against replay attacks.
func NewRequestExpiry(maxAge time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := r.Header.Get("X-Request-Timestamp")
			if raw == "" {
				http.Error(w, "missing X-Request-Timestamp", http.StatusBadRequest)
				return
			}
			unix, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				http.Error(w, "invalid X-Request-Timestamp", http.StatusBadRequest)
				return
			}
			ts := time.Unix(unix, 0)
			if time.Since(ts) > maxAge {
				http.Error(w, "request expired", http.StatusGone)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
