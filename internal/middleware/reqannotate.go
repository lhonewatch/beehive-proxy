package middleware

import (
	"net/http"
	"strings"
)

// AnnotateRule maps a header name to a static value injected into the request.
type AnnotateRule struct {
	Header string
	Value  string
}

// ParseAnnotateRules parses rules from a slice of "Header:Value" strings.
func ParseAnnotateRules(raw []string) []AnnotateRule {
	rules := make([]AnnotateRule, 0, len(raw))
	for _, r := range raw {
		parts := strings.SplitN(r, ":", 2)
		if len(parts) != 2 {
			continue
		}
		h := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		if h != "" {
			rules = append(rules, AnnotateRule{Header: h, Value: v})
		}
	}
	return rules
}

// NewRequestAnnotate returns middleware that injects static headers into every
// incoming request before passing it downstream. Existing headers are
// overwritten only when OverwriteExisting is true.
func NewRequestAnnotate(rules []AnnotateRule, overwrite bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, rule := range rules {
				if !overwrite && r.Header.Get(rule.Header) != "" {
					continue
				}
				r.Header.Set(rule.Header, rule.Value)
			}
			next.ServeHTTP(w, r)
		})
	}
}
