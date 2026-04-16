package config

import (
	"fmt"
	"net/http"

	"github.com/beehive-proxy/internal/middleware"
)

// JWTConfig holds JWT middleware configuration.
type JWTConfig struct {
	Enabled   bool
	Secret    []byte
	HeaderKey string
}

func jwtConfigFromEnv() (JWTConfig, error) {
	secret := envString("JWT_SECRET", "")
	if secret == "" {
		return JWTConfig{Enabled: false}, nil
	}
	headerKey := envString("JWT_SUBJECT_HEADER", "X-User-ID")
	return JWTConfig{
		Enabled:   true,
		Secret:    []byte(secret),
		HeaderKey: headerKey,
	}, nil
}

// JWTMiddleware returns a configured JWT middleware or nil when disabled.
func (c JWTConfig) JWTMiddleware() func(http.Handler) http.Handler {
	if !c.Enabled {
		return nil
	}
	return middleware.NewJWT(middleware.JWTOptions{
		Secret:    c.Secret,
		HeaderKey: c.HeaderKey,
	})
}

// Validate returns an error when JWT is enabled but misconfigured.
func (c JWTConfig) Validate() error {
	if c.Enabled && len(c.Secret) == 0 {
		return fmt.Errorf("JWT_SECRET must not be empty when JWT auth is enabled")
	}
	return nil
}
