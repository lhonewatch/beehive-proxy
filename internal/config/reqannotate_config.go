package config

import (
	"os"
	"strings"

	"github.com/example/beehive-proxy/internal/middleware"
)

func init() {
	registerMiddlewareBuilder(reqAnnotateMiddlewareFromEnv)
}

// reqAnnotateMiddlewareFromEnv reads:
//
//	BEEHIVE_ANNOTATE_RULES  — comma-separated "Header:Value" pairs
//	BEEHIVE_ANNOTATE_OVERWRITE — "true" to overwrite existing headers
func reqAnnotateMiddlewareFromEnv(cfg *Config) func(next interface{}) interface{} {
	raw := strings.TrimSpace(os.Getenv("BEEHIVE_ANNOTATE_RULES"))
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	rules := middleware.ParseAnnotateRules(parts)
	if len(rules) == 0 {
		return nil
	}

	overwrite := strings.EqualFold(strings.TrimSpace(os.Getenv("BEEHIVE_ANNOTATE_OVERWRITE")), "true")
	cfg.Middlewares = append(cfg.Middlewares, middleware.NewRequestAnnotate(rules, overwrite))
	return nil
}
