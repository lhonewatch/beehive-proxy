package middleware

import (
	"net/http"
	"strconv"
)

// SecurityHeadersOptions configures which security headers to inject.
type SecurityHeadersOptions struct {
	HSTSMaxAge            int
	FrameOptions          string
	ContentTypeNoSniff    bool
	XSSProtection         bool
	ReferrerPolicy        string
	PermissionsPolicy     string
}

// DefaultSecurityHeadersOptions returns sensible defaults.
func DefaultSecurityHeadersOptions() SecurityHeadersOptions {
	return SecurityHeadersOptions{
		HSTSMaxAge:         31536000,
		FrameOptions:       "SAMEORIGIN",
		ContentTypeNoSniff: true,
		XSSProtection:      true,
		ReferrerPolicy:     "strict-origin-when-cross-origin",
	}
}

// NewSecurityHeaders returns middleware that injects common security headers.
func NewSecurityHeaders(opts SecurityHeadersOptions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			if opts.HSTSMaxAge > 0 {
				h.Set("Strict-Transport-Security", "max-age="+strconv.Itoa(opts.HSTSMaxAge)+"; includeSubDomains")
			}
			if opts.FrameOptions != "" {
				h.Set("X-Frame-Options", opts.FrameOptions)
			}
			if opts.ContentTypeNoSniff {
				h.Set("X-Content-Type-Options", "nosniff")
			}
			if opts.XSSProtection {
				h.Set("X-XSS-Protection", "1; mode=block")
			}
			if opts.ReferrerPolicy != "" {
				h.Set("Referrer-Policy", opts.ReferrerPolicy)
			}
			if opts.PermissionsPolicy != "" {
				h.Set("Permissions-Policy", opts.PermissionsPolicy)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// NewRemoveHeaders returns middleware that removes the specified response
// headers before they are sent to the client. This is useful for stripping
// headers added by upstream services that should not be exposed externally
// (e.g. "X-Powered-By", "Server").
func NewRemoveHeaders(headers ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			h := w.Header()
			for _, name := range headers {
				h.Del(name)
			}
		})
	}
}
