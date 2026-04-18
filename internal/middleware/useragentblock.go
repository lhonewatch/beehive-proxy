package middleware

import (
	"net/http"
	"strings"
)

// UserAgentBlockOptions configures the user-agent blocking middleware.
type UserAgentBlockOptions struct {
	// Exact strings or substrings to block (case-insensitive).
	Patterns []string
	StatusCode int
}

// NewUserAgentBlock returns a middleware that rejects requests whose
// User-Agent header matches any of the configured patterns.
func NewUserAgentBlock(opts UserAgentBlockOptions, next http.Handler) http.Handler {
	if opts.StatusCode == 0 {
		opts.StatusCode = http.StatusForbidden
	}

	// Normalise patterns to lower-case once at construction time.
	lower := make([]string, len(opts.Patterns))
	for i, p := range opts.Patterns {
		lower[i] = strings.ToLower(p)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := strings.ToLower(r.Header.Get("User-Agent"))
		for _, p := range lower {
			if p != "" && strings.Contains(ua, p) {
				http.Error(w, http.StatusText(opts.StatusCode), opts.StatusCode)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
