package middleware

import (
	"net/http"
	"strings"
)

// NewRequestVersion injects an API version header into the request based on
// a path prefix match. E.g. /v1/ → "1", /v2/ → "2".
// The resolved version is written to versionHeader on the request before
// passing to the next handler. If no prefix matches, the header is set to
// defaultVersion.
func NewRequestVersion(versionHeader, defaultVersion string, prefixes map[string]string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		version := defaultVersion
		path := r.URL.Path
		for prefix, ver := range prefixes {
			if strings.HasPrefix(path, prefix) {
				version = ver
				break
			}
		}
		if versionHeader != "" && version != "" {
			r.Header.Set(versionHeader, version)
		}
		next.ServeHTTP(w, r)
	})
}
