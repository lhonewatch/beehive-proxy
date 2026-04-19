package config

import (
	"time"

	"github.com/example/beehive-proxy/internal/middleware"
)

func init() {
	configHooks = append(configHooks, reqExpiryMiddlewareFromEnv)
}

func reqExpiryMiddlewareFromEnv(cfg *Config) {
	enabled := envString("REQUEST_EXPIRY_ENABLED", "false")
	if enabled != "true" {
		return
	}
	maxAge := envDuration("REQUEST_EXPIRY_MAX_AGE", 5*time.Minute)
	cfg.Middlewares = append(cfg.Middlewares, middleware.NewRequestExpiry(maxAge))
}
