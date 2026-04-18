package config

import (
	"net/http"

	"github.com/example/beehive-proxy/internal/middleware"
)

func init() {
	registerMiddleware(reqPriorityMiddlewareFromEnv)
}

func reqPriorityMiddlewareFromEnv(cfg *Config) func(http.Handler) http.Handler {
	threshold := envInt("PRIORITY_REJECT_THRESHOLD", 0)
	if threshold == 0 {
		// feature disabled — still sets header but never rejects
	}
	return func(next http.Handler) http.Handler {
		return middleware.NewRequestPriority(next, threshold, nil)
	}
}
