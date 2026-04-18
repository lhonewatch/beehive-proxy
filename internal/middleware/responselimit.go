package middleware

import (
	"net/http"
)

// cappedResponseWriter wraps ResponseWriter and stops writing after maxBytes.
type cappedResponseWriter struct {
	http.ResponseWriter
	written int64
	max     int64
	capped  bool
}

func (c *cappedResponseWriter) Write(p []byte) (int, error) {
	if c.max <= 0 {
		return c.ResponseWriter.Write(p)
	}
	remaining := c.max - c.written
	if remaining <= 0 {
		c.capped = true
		return 0, nil
	}
	if int64(len(p)) > remaining {
		p = p[:remaining]
		c.capped = true
	}
	n, err := c.ResponseWriter.Write(p)
	c.written += int64(n)
	return n, err
}

// NewResponseSizeLimit returns middleware that truncates response bodies
// larger than maxBytes. Pass 0 to disable.
func NewResponseSizeLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if maxBytes <= 0 {
				next.ServeHTTP(w, r)
				return
			}
			cw := &cappedResponseWriter{ResponseWriter: w, max: maxBytes}
			next.ServeHTTP(cw, r)
			if cw.capped {
				// Header already sent; nothing more we can do except note truncation.
				w.Header().Set("X-Response-Truncated", "true")
			}
		})
	}
}
