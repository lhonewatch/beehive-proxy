package config

import (
	"net/http"

	"github.com/yourusername/beehive-proxy/internal/middleware"
)

func init() {
	registerMiddlewareBuilder(dedupeMiddlewareFromEnv)
}

// dedupeMiddlewareFromEnv returns a Dedupe middleware when DEDUPE_ENABLED=true.
func dedupeMiddlewareFromEnv(cfg *Config) (func(http.Handler) http.Handler, error) {
	if !envBool("DEDUPE_ENABLED", false) {
		return nil, nil
	}
	return middleware.NewDedupe(), nil
}

// envBool reads an environment variable as a boolean, returning def if unset or unparseable.
func envBool(key string, def bool) bool {
	v := envString(key, "")
	switch v {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return def
	}
}
