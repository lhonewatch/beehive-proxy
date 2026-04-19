package middleware

import (
	"net/http"
)

// NewRequestScheme injects an X-Request-Scheme header derived from the
// incoming request. It checks X-Forwarded-Proto first, then TLS state,
// and finally falls back to "http".
func NewRequestScheme(header string) func(http.Handler) http.Handler {
	if header == "" {
		header = "X-Request-Scheme"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scheme := detectScheme(r)
			r2 := r.Clone(r.Context())
			r2.Header.Set(header, scheme)
			next.ServeHTTP(w, r2)
		})
	}
}

func detectScheme(r *http.Request) string {
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return proto
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}
