package middleware

import (
	"net/http"
	"strings"
)

// PathBlockOptions configures which paths are blocked.
type PathBlockOptions struct {
	// ExactPaths are matched verbatim against r.URL.Path.
	ExactPaths []string
	// PrefixPaths block any request whose path starts with the given prefix.
	PrefixPaths []string
	StatusCode  int
}

// NewPathBlock returns middleware that returns StatusForbidden (or a
// configured status code) for any request matching a blocked path.
func NewPathBlock(opts PathBlockOptions) func(http.Handler) http.Handler {
	code := opts.StatusCode
	if code == 0 {
		code = http.StatusForbidden
	}

	exact := make(map[string]struct{}, len(opts.ExactPaths))
	for _, p := range opts.ExactPaths {
		exact[p] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			if _, ok := exact[path]; ok {
				http.Error(w, http.StatusText(code), code)
				return
			}

			for _, prefix := range opts.PrefixPaths {
				if strings.HasPrefix(path, prefix) {
					http.Error(w, http.StatusText(code), code)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
