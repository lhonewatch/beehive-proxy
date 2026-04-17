package config

import (
	"net/http"
	"os"
	"strings"

	"github.com/beehive-proxy/internal/middleware"
)

// StripPrefixConfig holds configuration for the strip-prefix middleware.
type StripPrefixConfig struct {
	Enabled bool
	Prefix  string
}

// stripPrefixConfigFromEnv reads strip-prefix settings from environment
// variables:
//
//	PROXY_STRIP_PREFIX        – the prefix to strip (enables middleware when set)
func stripPrefixConfigFromEnv() StripPrefixConfig {
	prefix := strings.TrimSpace(os.Getenv("PROXY_STRIP_PREFIX"))
	return StripPrefixConfig{
		Enabled: prefix != "",
		Prefix:  prefix,
	}
}

// Middleware returns an http.Handler middleware for strip-prefix, or nil when
// the feature is disabled.
func (c StripPrefixConfig) Middleware() func(http.Handler) http.Handler {
	if !c.Enabled {
		return nil
	}
	return middleware.NewStripPrefix(c.Prefix)
}
