package middleware

import (
	"net/http"
	"regexp"
	"strings"
)

// RewriteRule maps a compiled pattern to a replacement string for URL rewriting.
type QueryRewriteRule struct {
	Param       string
	Pattern     *regexp.Regexp
	Replacement string
}

// QueryTransformOptions configures the query parameter transform middleware.
type QueryTransformOptions struct {
	Rules []QueryRewriteRule
}

// NewQueryTransform returns middleware that rewrites query parameters based on rules.
func NewQueryTransform(opts QueryTransformOptions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(opts.Rules) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			q := r.URL.Query()
			modified := false

			for _, rule := range opts.Rules {
				vals, ok := q[rule.Param]
				if !ok {
					continue
				}
				for i, v := range vals {
					newVal := rule.Pattern.ReplaceAllString(v, rule.Replacement)
					if newVal != v {
						vals[i] = newVal
						modified = true
					}
				}
				q[rule.Param] = vals
			}

			if modified {
				r2 := r.Clone(r.Context())
				r2.URL.RawQuery = q.Encode()
				r = r2
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ParseQueryRewriteRules parses rules from strings in the form "param:pattern=replacement".
func ParseQueryRewriteRules(raw []string) ([]QueryRewriteRule, error) {
	rules := make([]QueryRewriteRule, 0, len(raw))
	for _, s := range raw {
		// format: param:pattern=replacement
		colon := strings.IndexByte(s, ':')
		if colon < 0 {
			continue
		}
		param := s[:colon]
		rest := s[colon+1:]
		eq := strings.IndexByte(rest, '=')
		if eq < 0 {
			continue
		}
		patStr := rest[:eq]
		replacement := rest[eq+1:]
		pat, err := regexp.Compile(patStr)
		if err != nil {
			return nil, err
		}
		rules = append(rules, QueryRewriteRule{Param: param, Pattern: pat, Replacement: replacement})
	}
	return rules, nil
}
