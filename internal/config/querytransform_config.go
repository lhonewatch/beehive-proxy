package config

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/beehive-proxy/internal/middleware"
)

// queryTransformConfigFromEnv builds a QueryTransform middleware from environment variables.
//
// QUERY_TRANSFORM_RULES — semicolon-separated list of rules in the form "param:pattern=replacement"
//
// Example:
//
//	QUERY_TRANSFORM_RULES="version:v1=v2;format:xml=json"
func queryTransformConfigFromEnv() (func(http.Handler) http.Handler, error) {
	raw := strings.TrimSpace(os.Getenv("QUERY_TRANSFORM_RULES"))
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ";")
	trimmed := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			trimmed = append(trimmed, p)
		}
	}

	if len(trimmed) == 0 {
		return nil, nil
	}

	rules, err := middleware.ParseQueryRewriteRules(trimmed)
	if err != nil {
		return nil, fmt.Errorf("QUERY_TRANSFORM_RULES: %w", err)
	}

	return middleware.NewQueryTransform(middleware.QueryTransformOptions{Rules: rules}), nil
}

func init() {
	registerMiddlewareLoader(func(cfg *Config) error {
		mw, err := queryTransformConfigFromEnv()
		if err != nil {
			return err
		}
		if mw != nil {
			cfg.Middlewares = append(cfg.Middlewares, mw)
		}
		return nil
	})
}
