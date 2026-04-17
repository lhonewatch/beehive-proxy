package middleware

import (
	"net/http"
	"strings"
)

// NewStripPrefix returns middleware that strips the given prefix from the
// request URL path before passing it to the next handler. If the request
// path does not start with the prefix the request is passed through
// unchanged.
func NewStripPrefix(prefix string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if prefix == "" {
				next.ServeHTTP(w, r)
				return
			}
			if !strings.HasPrefix(r.URL.Path, prefix) {
				next.ServeHTTP(w, r)
				return
			}
			stripped := strings.TrimPrefix(r.URL.Path, prefix)
			if stripped == "" {
				stripped = "/"
			}
			r2 := r.Clone(r.Context())
			r2.URL.Path = stripped
			if r.URL.RawPath != "" {
				r2.URL.RawPath = strings.TrimPrefix(r.URL.RawPath, prefix)
				if r2.URL.RawPath == "" {
					r2.URL.RawPath = "/"
				}
			}
			next.ServeHTTP(w, r2)
		})
	}
}
