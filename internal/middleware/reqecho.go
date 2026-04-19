package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

// NewRequestEcho injects selected request headers back into the response
// under a configurable prefix. Useful for debugging and tracing.
//
// Example: prefix="X-Echo", headers=["X-Trace-ID"] → response header
// "X-Echo-X-Trace-Id: <value>".
func NewRequestEcho(prefix string, headers []string) func(http.Handler) http.Handler {
	norm := make([]string, len(headers))
	for i, h := range headers {
		norm[i] = http.CanonicalHeaderKey(h)
	}
	if prefix == "" {
		prefix = "X-Echo"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			for _, h := range norm {
				vals := r.Header[h]
				if len(vals) == 0 {
					continue
				}
				key := prefix + "-" + strings.ReplaceAll(h, "-", "-")
				for i, v := range vals {
					if i == 0 {
						w.Header().Set(key, v)
					} else {
						w.Header().Add(key, v)
					}
				}
			}
			// Always echo the request method and path for convenience.
			w.Header().Set(prefix+"-Method", r.Method)
			w.Header().Set(prefix+"-Path", r.URL.Path)
			w.Header().Set(prefix+"-Proto", strconv.Itoa(r.ProtoMajor)+"."+strconv.Itoa(r.ProtoMinor))
		})
	}
}
