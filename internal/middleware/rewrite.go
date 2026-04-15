package middleware

import (
	"net/http"
	"regexp"
	"strings"
)

// RewriteRule defines a single URL rewrite rule using a regexp pattern and replacement.
type RewriteRule struct {
	Pattern     *regexp.Regexp
	Replacement string
}

// RewriteOptions configures the rewrite middleware.
type RewriteOptions struct {
	Rules []RewriteRule
	// StripPrefix removes a leading prefix from the request path before proxying.
	StripPrefix string
}

// NewRewrite returns middleware that rewrites request paths before passing them
// to the next handler. Rules are applied in order; the first matching rule wins.
// If StripPrefix is set it is applied after rule matching.
func NewRewrite(opts RewriteOptions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			for _, rule := range opts.Rules {
				if rule.Pattern.MatchString(path) {
					path = rule.Pattern.ReplaceAllString(path, rule.Replacement)
					break
				}
			}

			if opts.StripPrefix != "" {
				path = strings.TrimPrefix(path, opts.StripPrefix)
				if path == "" {
					path = "/"
				}
			}

			// Clone the request so we don't mutate the original.
			modified := r.Clone(r.Context())
			modified.URL.Path = path
			modified.RequestURI = path
			if modified.URL.RawQuery != "" {
				modified.RequestURI += "?" + modified.URL.RawQuery
			}

			next.ServeHTTP(w, modified)
		})
	}
}
