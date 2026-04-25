package config

import (
	"time"

	"github.com/beehive-proxy/internal/middleware"
)

func init() {
	registerMiddlewareBuilder(reqIdempotencyMiddlewareFromEnv)
}

// reqIdempotencyMiddlewareFromEnv reads idempotency settings from the environment.
//
//	BEEHIVE_IDEMPOTENCY_ENABLED   – "true" to enable (default: false)
//	BEEHIVE_IDEMPOTENCY_HEADER    – request header carrying the key (default: "Idempotency-Key")
//	BEEHIVE_IDEMPOTENCY_TTL       – cache TTL as a Go duration (default: 5m)
func reqIdempotencyMiddlewareFromEnv(cfg *Config) {
	if !envBool("BEEHIVE_IDEMPOTENCY_ENABLED", false) {
		return
	}

	header := envString("BEEHIVE_IDEMPOTENCY_HEADER", "Idempotency-Key")
	ttl := envDuration("BEEHIVE_IDEMPOTENCY_TTL", 5*time.Minute)

	cfg.Middlewares = append(cfg.Middlewares, middleware.NewRequestIdempotency(header, ttl))
}
