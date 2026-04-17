package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const RequestIDHeader = "X-Request-ID"

// NewRequestID injects a unique request ID into each request and response.
// If the incoming request already carries a X-Request-ID header it is reused.
func NewRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if id == "" {
			id = generateRequestID()
		}

		// Propagate to downstream service.
		r.Header.Set(RequestIDHeader, id)
		// Expose on the response so callers can correlate.
		w.Header().Set(RequestIDHeader, id)

		next.ServeHTTP(w, r)
	})
}

func generateRequestID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback: all-zeros id is still better than nothing.
		return "0000000000000000"
	}
	return hex.EncodeToString(b)
}
