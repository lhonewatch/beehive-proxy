package config

import (
	"net/http"
	"time"

	"github.com/sixafter/beehive-proxy/internal/middleware"
)

func init() {
	registerMiddlewareLoader(reqCacheMiddlewareFromEnv)
}

// reqCacheMiddlewareFromEnv reads REQUEST_CACHE_ENABLED and
// REQUEST_CACHE_TTL from the environment and, when enabled, returns
// a middleware that caches upstream GET responses.
func reqCacheMiddlewareFromEnv(env map[string]string) func(http.Handler) http.Handler {
	enabled := envStringMap(env, "REQUEST_CACHE_ENABLED", "false")
	if enabled != "true" {
		return nil
	}

	ttlStr := envStringMap(env, "REQUEST_CACHE_TTL", "30s")
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil || ttl <= 0 {
		ttl = 30 * time.Second
	}

	rc := middleware.NewRequestCache(ttl)
	return rc.Handler
}

// envStringMap is a helper used by sub-config loaders to read from a
// pre-built env map (used in tests and in FromEnv).
func envStringMap(env map[string]string, key, fallback string) string {
	if v, ok := env[key]; ok && v != "" {
		return v
	}
	return fallback
}
