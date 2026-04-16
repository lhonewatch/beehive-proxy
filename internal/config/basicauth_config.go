package config

import (
	"os"
	"strings"

	"github.com/beehive-proxy/internal/middleware"
)

// BasicAuthConfig holds parsed basic-auth settings.
type BasicAuthConfig struct {
	Enabled     bool
	Realm       string
	Credentials map[string]string
}

// basicAuthConfigFromEnv reads basic-auth configuration from environment
// variables:
//
//	BASIC_AUTH_ENABLED  – "true" to enable (default: false)
//	BASIC_AUTH_REALM    – realm string (default: "beehive-proxy")
//	BASIC_AUTH_USERS    – comma-separated user:password pairs
func basicAuthConfigFromEnv() BasicAuthConfig {
	cfg := BasicAuthConfig{
		Enabled:     strings.EqualFold(os.Getenv("BASIC_AUTH_ENABLED"), "true"),
		Realm:       envString("BASIC_AUTH_REALM", "beehive-proxy"),
		Credentials: make(map[string]string),
	}
	raw := os.Getenv("BASIC_AUTH_USERS")
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			cfg.Credentials[parts[0]] = parts[1]
		}
	}
	return cfg
}

// BasicAuthMiddleware returns a configured BasicAuth middleware or nil when
// basic auth is disabled.
func (c BasicAuthConfig) BasicAuthMiddleware() func(http.Handler) http.Handler {
	if !c.Enabled {
		return nil
	}
	return middleware.NewBasicAuth(middleware.BasicAuthOptions{
		Credentials: c.Credentials,
		Realm:       c.Realm,
	})
}
