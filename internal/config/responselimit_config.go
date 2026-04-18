package config

import (
	"net/http"

	"github.com/beehive-proxy/internal/middleware"
)

func init() {
	registerMiddlewareBuilder(responseLimitMiddlewareFromEnv)
}

// responseLimitMiddlewareFromEnv reads RESPONSE_SIZE_LIMIT_BYTES (default 0 =
// disabled) and returns a ResponseSizeLimit middleware when enabled.
func responseLimitMiddlewareFromEnv(cfg *Config) func(http.Handler) http.Handler {
	maxBytes := envInt("RESPONSE_SIZE_LIMIT_BYTES", 0)
	if maxBytes <= 0 {
		return nil
	}
	cfg.ResponseSizeLimitBytes = int64(maxBytes)
	return middleware.NewResponseSizeLimit(int64(maxBytes))
}
