package config

import (
	"net/http"
	"strings"

	"github.com/beehive-proxy/internal/middleware"
)

func init() {
	configLoaders = append(configLoaders, trustedProxyConfigFromEnv)
}

type TrustedProxyConfig struct {
	Enabled bool
	CIDRs   []string
}

func trustedProxyConfigFromEnv(cfg *Config) {
	enabled := envString("TRUSTED_PROXY_ENABLED", "false")
	if enabled != "true" {
		return
	}
	raw := envString("TRUSTED_PROXY_CIDRS", "")
	var cidrs []string
	for _, c := range strings.Split(raw, ",") {
		c = strings.TrimSpace(c)
		if c != "" {
			cidrs = append(cidrs, c)
		}
	}
	if len(cidrs) == 0 {
		return
	}
	m := middleware.NewTrustedProxy(middleware.TrustedProxyOptions{TrustedCIDRs: cidrs})
	cfg.Middlewares = append(cfg.Middlewares, func(next http.Handler) http.Handler {
		return m(next)
	})
}
