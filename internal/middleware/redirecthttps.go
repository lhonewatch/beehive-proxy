package middleware

import (
	"net/http"
)

// NewRedirectHTTPS returns middleware that redirects HTTP requests to HTTPS.
// If behindProxy is true, it checks X-Forwarded-Proto before redirecting.
func NewRedirectHTTPS(behindProxy bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isHTTPS(r, behindProxy) {
				next.ServeHTTP(w, r)
				return
			}
			target := "https://" + r.Host + r.URL.RequestURI()
			http.Redirect(w, r, target, http.StatusMovedPermanently)
		})
	}
}

func isHTTPS(r *http.Request, behindProxy bool) bool {
	if r.TLS != nil {
		return true
	}
	if behindProxy {
		return r.Header.Get("X-Forwarded-Proto") == "https"
	}
	return false
}
