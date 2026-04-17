package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// AccessLogOptions configures the access log middleware.
type AccessLogOptions struct {
	Logger       *zap.Logger
	SkipPaths    []string
}

// NewAccessLog returns middleware that logs each request in a structured
// Apache-style combined-log format using zap.
func NewAccessLog(opts AccessLogOptions) func(http.Handler) http.Handler {
	skip := make(map[string]struct{}, len(opts.SkipPaths))
	for _, p := range opts.SkipPaths {
		skip[p] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := skip[r.URL.Path]; ok {
				next.ServeHTTP(w, r)
				return
			}

			rec := NewResponseRecorder(w)
			start := time.Now()
			next.ServeHTTP(rec, r)
			dur := time.Since(start)

			opts.Logger.Info("access",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("query", r.URL.RawQuery),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
				zap.Int("status", rec.Status()),
				zap.Int("bytes", rec.BytesWritten()),
				zap.Duration("duration", dur),
				zap.String("request_id", r.Header.Get("X-Request-Id")),
			)
		})
	}
}
