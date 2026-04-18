package config

import (
	"net/http"
	"strings"

	"github.com/beehive-proxy/internal/middleware"
)

func init() {
	registerMiddlewareBuilder(reqFingerprintMiddlewareFromEnv)
}

// reqFingerprintMiddlewareFromEnv reads:
//   FINGERPRINT_ENABLED=true
//   FINGERPRINT_HEADERS=X-Tenant,X-App-ID   (comma-separated, optional)
func reqFingerprintMiddlewareFromEnv(next http.Handler) http.Handler {
	if envString("FINGERPRINT_ENABLED", "false") != "true" {
		return next
	}
	raw := envString("FINGERPRINT_HEADERS", "")
	var headers []string
	for _, h := range strings.Split(raw, ",") {
		h = strings.TrimSpace(h)
		if h != "" {
			headers = append(headers, h)
		}
	}
	return middleware.NewRequestFingerprint(headers, next)
}
