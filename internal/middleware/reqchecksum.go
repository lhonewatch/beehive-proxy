package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

// NewRequestChecksum computes a SHA-256 checksum of the incoming request body
// and injects it as a request header before passing to the next handler.
// The body is fully buffered so downstream handlers can still read it.
//
// headerName defaults to "X-Request-Checksum" when empty.
func NewRequestChecksum(next http.Handler, headerName string) http.Handler {
	if headerName == "" {
		headerName = "X-Request-Checksum"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil || r.Body == http.NoBody {
			r.Header.Set(headerName, "")
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}
		_ = r.Body.Close()

		sum := sha256.Sum256(body)
		r.Header.Set(headerName, hex.EncodeToString(sum[:]))
		r.Body = io.NopCloser(bytes.NewReader(body))
		r.ContentLength = int64(len(body))

		next.ServeHTTP(w, r)
	})
}
