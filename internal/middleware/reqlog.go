package middleware

import (
	"bytes"
	"io"
	"net/http"
)

// RequestBodyLogger captures the request body for inspection and restores it
// so downstream handlers can still read it.
type RequestBodyLogger struct {
	OnBody  func(body []byte, r *http.Request)
	MaxSize int64
}

// NewRequestBodyLogger returns a middleware that reads up to MaxSize bytes of
// the request body, calls OnBody, then restores the body for downstream use.
// If MaxSize is 0 a default of 4096 bytes is used.
//
// Note: if the request body exceeds MaxSize, only the first MaxSize bytes are
// passed to OnBody, but the full original content is still forwarded to the
// next handler.
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

			// Read up to maxSize bytes, then capture any remaining bytes so
			// the full body can be restored for downstream handlers.
			limited := io.LimitReader(r.Body, maxSize)
			buf, err := io.ReadAll(limited)
			if err != nil {
				_ = r.Body.Close()
				http.Error(w, "failed to read body", http.StatusInternalServerError)
				return
			}

			// Capture any bytes beyond maxSize so we can restore the full body.
			remainder, err := io.ReadAll(r.Body)
			_ = r.Body.Close()
			if err != nil {
				http.Error(w, "failed to read body", http.StatusInternalServerError)
				return
			}

			if onBody != nil {
				onBody(buf, r)
			}

			// Restore the complete body (sampled portion + remainder) for
			// downstream handlers.
			r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(buf), bytes.NewReader(remainder)))
			next.ServeHTTP(w, r)
		})
	}
}
