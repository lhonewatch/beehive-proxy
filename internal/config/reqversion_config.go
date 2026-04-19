package config

import (
	"net/http"
	"os"
	"strings"

	"github.com/beehive-proxy/internal/middleware"
)

// reqVersionConfigFromEnv reads:
//   PROXY_API_VERSION_HEADER  — header name (default: X-API-Version)
//   PROXY_API_VERSION_DEFAULT — fallback version string (default: "")
//   PROXY_API_VERSION_PREFIXES — comma-separated prefix=version pairs, e.g. /v1/=1,/v2/=2
func reqVersionConfigFromEnv() func(http.Handler) http.Handler {
	header := envString("PROXY_API_VERSION_HEADER", "X-API-Version")
	defaultVer := envString("PROXY_API_VERSION_DEFAULT", "")
	raw := os.Getenv("PROXY_API_VERSION_PREFIXES")
	if raw == "" {
		return nil
	}
	prefixes := map[string]string{}
	for _, pair := range strings.Split(raw, ",") {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 && parts[0] != "" {
			prefixes[parts[0]] = parts[1]
		}
	}
	if len(prefixes) == 0 {
		return nil
	}
	return func(next http.Handler) http.Handler {
		return middleware.NewRequestVersion(header, defaultVer, prefixes, next)
	}
}

func init() {
	registerMiddlewareFactory(reqVersionConfigFromEnv)
}
