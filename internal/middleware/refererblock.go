package middleware

import (
	"net/http"
	"strings"
)

// NewRefererBlock returns middleware that blocks requests whose Referer header
// matches any of the provided patterns (exact or prefix match).
func NewRefererBlock(blocked []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(blocked) > 0 {
				referer := r.Header.Get("Referer")
				if referer != "" {
					for _, pattern := range blocked {
						if matchReferer(referer, pattern) {
							http.Error(w, "Forbidden", http.StatusForbidden)
							return
						}
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func matchReferer(referer, pattern string) bool {
	referer = strings.ToLower(referer)
	pattern = strings.ToLower(pattern)
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(referer, strings.TrimSuffix(pattern, "*"))
	}
	return referer == pattern
}
