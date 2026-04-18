package config

import (
	"log/slog"
	"net/http"

	"github.com/beehive-proxy/internal/middleware"
)

type ShadowConfig struct {
	Enabled   bool
	TargetURL string
}

func shadowConfigFromEnv() ShadowConfig {
	return ShadowConfig{
		Enabled:   envString("SHADOW_ENABLED", "false") == "true",
		TargetURL: envString("SHADOW_TARGET_URL", ""),
	}
}

func init() {
	registerMiddlewareBuilder(func(cfg *Config, logger *slog.Logger) (func(http.Handler) http.Handler, error) {
		sc := shadowConfigFromEnv()
		if !sc.Enabled || sc.TargetURL == "" {
			return nil, nil
		}
		return middleware.NewShadow(sc.TargetURL, logger)
	})
}
