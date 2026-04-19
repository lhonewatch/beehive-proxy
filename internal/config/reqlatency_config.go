package config

import (
	"time"

	"github.com/yourusername/beehive-proxy/internal/middleware"
)

type ReqLatencyConfig struct {
	Enabled   bool
	Threshold time.Duration
}

func reqLatencyConfigFromEnv() ReqLatencyConfig {
	enabled := envString("LATENCY_HEADER_ENABLED", "false") == "true"
	threshold := envDuration("LATENCY_SLOW_THRESHOLD", 0)
	return ReqLatencyConfig{
		Enabled:   enabled,
		Threshold: threshold,
	}
}

func init() {
	registerMiddlewareBuilder(func(cfg *Config) (mwEntry, bool) {
		c := reqLatencyConfigFromEnv()
		if !c.Enabled {
			return mwEntry{}, false
		}
		return mwEntry{
			Name: "reqlatency",
			MW:   middleware.NewRequestLatency(c.Threshold),
		}, true
	})
}
