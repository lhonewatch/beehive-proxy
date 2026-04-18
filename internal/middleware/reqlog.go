package middleware

import (
	"bytes"
	"io"
	"net/http"
)

// RequestBodyLogger captures the request body for inspection and restores it
// so downstream handlers can still read it.
type RequestBodyLogger struct {
	OnBody func(body []byte, r *http.Request)
	MaxSize int64
}

// NewRequestBodyLogger returns a middleware that reads up to MaxSize bytes of
// the request body, calls OnBody, then restores the body for downstream use.
// If MaxSize is 0 a default of 4096 bytes is used.
func NewRequestBodyLogger(maxSize int64, onBody func([]byte, *http.Request)) func(http.Handler) http.Handler {
	if maxSize <= 0 {
		maxSize = 4096
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body == nil || r.Body == http.NoBody {
				next.ServeHTTP(w, r)
				return
			}

			buf, err := io.ReadAll(io.LimitReader(r.Body, maxSize))
			_ = r.Body.Close()
			if err != nil {
				http.Error(w, "failed to read body", http.StatusInternalServerError)
				return
			}

			if onBody != nil {
				onBody(buf, r)
			}

			r.Body = io.NopCloser(bytes.NewReader(buf))
			next.ServeHTTP(w, r)
		})
	}
}
