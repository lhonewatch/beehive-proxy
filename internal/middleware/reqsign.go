package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"
)

// NewRequestSigning adds an HMAC-SHA256 signature header to outgoing requests.
// The signature covers: method + path + timestamp (rounded to the minute).
func NewRequestSigning(secret, signatureHeader, timestampHeader string) func(http.Handler) http.Handler {
	if signatureHeader == "" {
		signatureHeader = "X-Signature"
	}
	if timestampHeader == "" {
		timestampHeader = "X-Timestamp"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ts := r.Header.Get(timestampHeader)
			if ts == "" {
				ts = time.Now().UTC().Format(time.RFC3339)
				r.Header.Set(timestampHeader, ts)
			}
			sig := sign(secret, r.Method, r.URL.Path, ts)
			r.Header.Set(signatureHeader, sig)
			next.ServeHTTP(w, r)
		})
	}
}

func sign(secret, method, path, ts string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(method + "\n" + path + "\n" + ts))
	return hex.EncodeToString(mac.Sum(nil))
}
