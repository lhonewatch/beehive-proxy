package middleware

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
)

// NewRequestFingerprint adds an X-Request-Fingerprint header derived from
// selected request attributes (method, path, selected headers).
// The fingerprint can be used for deduplication, caching keys, or audit logs.
func NewRequestFingerprint(hashHeaders []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fp := fingerprint(r, hashHeaders)
		r.Header.Set("X-Request-Fingerprint", fp)
		next.ServeHTTP(w, r)
	})
}

func fingerprint(r *http.Request, extraHeaders []string) string {
	var b strings.Builder
	b.WriteString(r.Method)
	b.WriteByte('|')
	b.WriteString(r.URL.Path)
	b.WriteByte('|')
	b.WriteString(r.URL.RawQuery)
	for _, h := range extraHeaders {
		b.WriteByte('|')
		b.WriteString(strings.ToLower(h))
		b.WriteByte('=')
		b.WriteString(r.Header.Get(h))
	}
	sum := sha256.Sum256([]byte(b.String()))
	return fmt.Sprintf("%x", sum[:8])
}
