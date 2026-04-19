package middleware

import (
	"net/http"
	"strings"
)

// TagRule maps a path prefix to a tag value.
type TagRule struct {
	Prefix string
	Tag    string
}

// NewRequestTag injects an X-Request-Tag header based on the first matching
// path prefix rule. If no rule matches, the request passes through unmodified.
func NewRequestTag(rules []TagRule, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, rule := range rules {
			if rule.Prefix != "" && strings.HasPrefix(r.URL.Path, rule.Prefix) {
				r.Header.Set("X-Request-Tag", rule.Tag)
				break
			}
		}
		next.ServeHTTP(w, r)
	})
}

// ParseTagRules parses a semicolon-separated list of "prefix=tag" pairs.
// Example: "/api=api;/static=static"
func ParseTagRules(raw string) []TagRule {
	var rules []TagRule
	for _, part := range strings.Split(raw, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		idx := strings.IndexByte(part, '=')
		if idx <= 0 {
			continue
		}
		rules = append(rules, TagRule{
			Prefix: strings.TrimSpace(part[:idx]),
			Tag:    strings.TrimSpace(part[idx+1:]),
		})
	}
	return rules
}
