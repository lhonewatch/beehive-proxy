package config

import (
	"net/http"
	"time"

	"github.com/beehive-proxy/internal/middleware"
)

type ReqCloneConfig struct {
	Enabled   bool
	TargetURL string
	Timeout   time.Duration
}

func reqCloneConfigFromEnv() ReqCloneConfig {
	target := envString("CLONE_TARGET_URL", "")
	return ReqCloneConfig{
		Enabled:   target != "",
		TargetURL: target,
		Timeout:   envDuration("CLONE_TIMEOUT", 5*time.Second),
	}
}

func init() {
	registerMiddlewareBuilder(func(cfg *Config) func(http.Handler) http.Handler {
		c := reqCloneConfigFromEnv()
		if !c.Enabled {
			return nil
		}
		client := &http.Client{Timeout: c.Timeout}
		return middleware.NewRequestClone(c.TargetURL, client)
	})
}
