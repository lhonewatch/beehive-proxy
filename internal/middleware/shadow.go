package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

// NewShadow returns a middleware that clones each incoming request and sends it
// to shadowURL asynchronously, discarding the response. The primary request is
// unaffected.
func NewShadow(shadowURL string, logger *slog.Logger) (func(http.Handler) http.Handler, error) {
	target, err := url.ParseRequestURI(shadowURL)
	if err != nil {
		return nil, err
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			go shadowRequest(r, target, logger)
			next.ServeHTTP(w, r)
		})
	}, nil
}

func shadowRequest(r *http.Request, target *url.URL, logger *slog.Logger) {
	var body []byte
	if r.Body != nil {
		var err error
		body, err = io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			return
		}
	}

	cloned, err := http.NewRequest(r.Method, target.String()+r.RequestURI, bytes.NewReader(body))
	if err != nil {
		return
	}
	for k, v := range r.Header {
		cloned.Header[k] = v
	}
	cloned.Header.Set("X-Shadow-Request", "1")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(cloned)
	if err != nil {
		if logger != nil {
			logger.Warn("shadow request failed", "error", err)
		}
		return
	}
	resp.Body.Close()
}
