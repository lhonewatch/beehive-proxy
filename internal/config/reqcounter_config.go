package config

import (
	"net/http"

	"github.com/yourorg/beehive-proxy/internal/middleware"
)

// RequestCounterConfig holds settings for the request counter middleware.
type RequestCounterConfig struct {
	Enabled bool
}

func reqCounterConfigFromEnv() RequestCounterConfig {
	return RequestCounterConfig{
		Enabled: envString("REQUEST_COUNTER_ENABLED", "true") == "true",
	}
}

func init() {
	registerMiddlewareBuilder(func(cfg *Config) func(http.Handler) http.Handler {
		rc := reqCounterConfigFromEnv()
		if !rc.Enabled {
			return nil
		}
		return func(next http.Handler) http.Handler {
			return middleware.NewRequestCounter(next)
		}
	})
}
