package middleware

import (
	"net/http"
	"time"
)

const RequestTimestampHeader = "X-Request-Timestamp"

// NewRequestTimestamp injects an X-Request-Timestamp header (RFC3339Nano) into
// every incoming request so downstream handlers and logs can reference the
// exact moment the proxy received the request.
func NewRequestTimestamp(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(RequestTimestampHeader) == "" {
			r.Header.Set(RequestTimestampHeader, time.Now().UTC().Format(time.RFC3339Nano))
		}
		next.ServeHTTP(w, r)
	})
}
