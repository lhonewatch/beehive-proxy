package middleware

import (
	"net/http"
	"strings"
)

// NewMethodOverride allows clients to override the HTTP method via a header or
// query parameter. Useful for clients that only support GET/POST.
//
// Priority order:
//  1. X-HTTP-Method-Override header
//  2. X-Method-Override header
//  3. _method query parameter
func NewMethodOverride(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			override := r.Header.Get("X-HTTP-Method-Override")
			if override == "" {
				override = r.Header.Get("X-Method-Override")
			}
			if override == "" {
				override = r.URL.Query().Get("_method")
			}
			if override != "" {
				overrideUp := strings.ToUpper(override)
				allowed := map[string]bool{
					http.MethodPut:    true,
					http.MethodPatch:  true,
					http.MethodDelete: true,
				}
				if allowed[overrideUp] {
					r.Method = overrideUp
					r.Header.Del("X-HTTP-Method-Override")
					r.Header.Del("X-Method-Override")
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
