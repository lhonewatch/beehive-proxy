package config

import (
	"net/http"

	"github.com/andrebq/beehive-proxy/internal/middleware"
)

func reqTagMiddlewareFromEnv(cfg *Config) {
	raw := envString("REQTAG_RULES", "")
	if raw == "" {
		return
	}
	rules := middleware.ParseTagRules(raw)
	if len(rules) == 0 {
		return
	}
	cfg.Middlewares = append(cfg.Middlewares, func(next http.Handler) http.Handler {
		return middleware.NewRequestTag(rules, next)
	})
}

func init() {
	registerConfigLoader(reqTagMiddlewareFromEnv)
}
