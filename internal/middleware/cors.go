package middleware

import (
	"net/http"
	"strings"
)

// CORSOptions holds configuration for the CORS middleware.
type CORSOptions struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAge         string
}

// DefaultCORSOptions returns permissive defaults suitable for development.
func DefaultCORSOptions() CORSOptions {
	return CORSOptions{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-Trace-ID"},
		MaxAge:         "86400",
	}
}

// NewCORS returns a middleware that sets CORS headers on every response.
// Preflight OPTIONS requests are answered immediately with 204 No Content.
func NewCORS(opts CORSOptions) func(http.Handler) http.Handler {
	allowedOrigins := make(map[string]struct{}, len(opts.AllowedOrigins))
	for _, o := range opts.AllowedOrigins {
		allowedOrigins[o] = struct{}{}
	}

	methods := strings.Join(opts.AllowedMethods, ", ")
	headers := strings.Join(opts.AllowedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			allowed := false
			if _, ok := allowedOrigins["*"]; ok {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				allowed = true
			} else if _, ok := allowedOrigins[origin]; ok && origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Add("Vary", "Origin")
				allowed = true
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Methods", methods)
				w.Header().Set("Access-Control-Allow-Headers", headers)
				if opts.MaxAge != "" {
					w.Header().Set("Access-Control-Max-Age", opts.MaxAge)
				}
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
