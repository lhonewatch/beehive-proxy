package config

import (
	"net/http"

	"github.com/yourusername/beehive-proxy/internal/middleware"
)

type reqThrottleConfig struct {
	enabled bool
	rate    float64
	burst   float64
}

func reqThrottleConfigFromEnv() reqThrottleConfig {
	enabled := envString("THROTTLE_ENABLED", "false") == "true"
	rate := float64(envInt("THROTTLE_RATE", 100))
	burst := float64(envInt("THROTTLE_BURST", 200))
	return reqThrottleConfig{enabled: enabled, rate: rate, burst: burst}
}

func init() {
	registerMiddlewareBuilder(func(cfg *Config, next http.Handler) http.Handler {
		tc := reqThrottleConfigFromEnv()
		if !tc.enabled {
			return next
		}
		return middleware.NewRequestThrottle(tc.rate, tc.burst, next)
	})
}
