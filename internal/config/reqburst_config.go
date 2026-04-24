package config

import (
	"fmt"
	"net/http"
	"time"

	"github.com/beehive-proxy/internal/middleware"
)

type reqBurstConfig struct {
	Enabled bool
	Limit   int64
	Window  time.Duration
}

func reqBurstConfigFromEnv() reqBurstConfig {
	limit := int64(envInt("BURST_LIMIT", 0))
	window := envDuration("BURST_WINDOW", 10*time.Second)
	return reqBurstConfig{
		Enabled: limit > 0,
		Limit:   limit,
		Window:  window,
	}
}

func init() {
	registerMiddlewareBuilder(func(cfg *Config) (func(http.Handler) http.Handler, error) {
		bc := reqBurstConfigFromEnv()
		if !bc.Enabled {
			return nil, nil
		}
		if bc.Window <= 0 {
			return nil, fmt.Errorf("BURST_WINDOW must be positive")
		}
		return middleware.NewRequestBurst(bc.Limit, bc.Window), nil
	})
}
