package middleware

import (
	"bytes"
	"io"
	"net/http"
)

// TransformFunc receives the request body bytes and returns transformed bytes.
type TransformFunc func([]byte) ([]byte, error)

// NewBodyTransform returns middleware that applies fn to every request body.
// Requests with no body are passed through unchanged.
func NewBodyTransform(fn TransformFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body == nil || r.ContentLength == 0 {
				next.ServeHTTP(w, r)
				return
			}

			orig, err := io.ReadAll(r.Body)
			_ = r.Body.Close()
			if err != nil {
				http.Error(w, "failed to read body", http.StatusBadRequest)
				return
			}

			transformed, err := fn(orig)
			if err != nil {
				http.Error(w, "body transform error", http.StatusBadRequest)
				return
			}

			r2 := r.Clone(r.Context())
			r2.Body = io.NopCloser(bytes.NewReader(transformed))
			r2.ContentLength = int64(len(transformed))
			next.ServeHTTP(w, r2)
		})
	}
}
