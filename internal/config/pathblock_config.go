package config

import (
	"net/http"
	"strings"

	"github.com/beehive-proxy/internal/middleware"
)

func init() {
	registerMiddlewareBuilder(pathBlockMiddlewareFromEnv)
}

// pathBlockMiddlewareFromEnv reads BLOCK_EXACT_PATHS and BLOCK_PREFIX_PATHS
// (comma-separated) and BLOCK_STATUS_CODE from the environment and, when at
// least one path is configured, appends a PathBlock middleware to cfg.
//
// Example:
//
//	BLOCK_EXACT_PATHS=/secret,/internal
//	BLOCK_PREFIX_PATHS=/admin,/debug
//	BLOCK_STATUS_CODE=404
func pathBlockMiddlewareFromEnv(cfg *Config) {
	exactRaw := envString("BLOCK_EXACT_PATHS", "")
	prefixRaw := envString("BLOCK_PREFIX_PATHS", "")

	if exactRaw == "" && prefixRaw == "" {
		return
	}

	var exact, prefix []string
	for _, p := range strings.Split(exactRaw, ",") {
		if t := strings.TrimSpace(p); t != "" {
			exact = append(exact, t)
		}
	}
	for _, p := range strings.Split(prefixRaw, ",") {
		if t := strings.TrimSpace(p); t != "" {
			prefix = append(prefix, t)
		}
	}

	code := envInt("BLOCK_STATUS_CODE", http.StatusForbidden)

	opts := middleware.PathBlockOptions{
		ExactPaths:  exact,
		PrefixPaths: prefix,
		StatusCode:  code,
	}
	cfg.Middlewares = append(cfg.Middlewares, middleware.NewPathBlock(opts))
}
