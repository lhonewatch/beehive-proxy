package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const (
	DefaultCorrelationHeader = "X-Correlation-ID"
)

// NewRequestCorrelation ensures every request carries a correlation ID.
// If the upstream already sends the header it is preserved; otherwise a new
// random ID is generated. The ID is also echoed back in the response.
func NewRequestCorrelation(header string, next http.Handler) http.Handler {
	if header == "" {
		header = DefaultCorrelationHeader
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(header)
		if id == "" {
			id = generateCorrelationID()
			r.Header.Set(header, id)
		}
		w.Header().Set(header, id)
		next.ServeHTTP(w, r)
	})
}

func generateCorrelationID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
