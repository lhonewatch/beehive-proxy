package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"
)

const (
	DefaultWatermarkHeader    = "X-Watermark"
	DefaultWatermarkTSHeader  = "X-Watermark-Ts"
)

// NewRequestWatermark stamps each request with an HMAC-SHA256 watermark
// derived from the secret, the request method, path, and a UTC timestamp.
// The timestamp is written to WatermarkTSHeader so downstream services can
// verify freshness independently.
func NewRequestWatermark(secret, header, tsHeader string) func(http.Handler) http.Handler {
	if header == "" {
		header = DefaultWatermarkHeader
	}
	if tsHeader == "" {
		tsHeader = DefaultWatermarkTSHeader
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ts := r.Header.Get(tsHeader)
			if ts == "" {
				ts = time.Now().UTC().Format(time.RFC3339)
				r.Header.Set(tsHeader, ts)
			}
			wm := watermark(secret, r.Method, r.URL.Path, ts)
			r.Header.Set(header, wm)
			next.ServeHTTP(w, r)
		})
	}
}

func watermark(secret, method, path, ts string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(method))
	mac.Write([]byte(":"))
	mac.Write([]byte(path))
	mac.Write([]byte(":"))
	mac.Write([]byte(ts))
	return hex.EncodeToString(mac.Sum(nil))
}
