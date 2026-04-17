package config

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/andrebq/beehive-proxy/internal/middleware"
)

// BodyTransformConfig holds configuration for the body transform middleware.
type BodyTransformConfig struct {
	Enabled bool
	// Mode: "uppercase", "lowercase", "base64encode", "base64decode", or "none"
	Mode string
}

func bodyTransformConfigFromEnv() BodyTransformConfig {
	mode := envString("BODY_TRANSFORM_MODE", "none")
	return BodyTransformConfig{
		Enabled: mode != "none" && mode != "",
		Mode:    mode,
	}
}

// Middleware returns the configured middleware or nil when disabled.
func (c BodyTransformConfig) Middleware() func(http.Handler) http.Handler {
	if !c.Enabled {
		return nil
	}
	switch strings.ToLower(c.Mode) {
	case "uppercase":
		return middleware.NewBodyTransform(func(b []byte) ([]byte, error) {
			return []byte(strings.ToUpper(string(b))), nil
		})
	case "lowercase":
		return middleware.NewBodyTransform(func(b []byte) ([]byte, error) {
			return []byte(strings.ToLower(string(b))), nil
		})
	case "base64encode":
		return middleware.NewBodyTransform(func(b []byte) ([]byte, error) {
			dst := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
			base64.StdEncoding.Encode(dst, b)
			return dst, nil
		})
	case "base64decode":
		return middleware.NewBodyTransform(func(b []byte) ([]byte, error) {
			dst := make([]byte, base64.StdEncoding.DecodedLen(len(b)))
			n, err := base64.StdEncoding.Decode(dst, b)
			if err != nil {
				return nil, err
			}
			return dst[:n], nil
		})
	}
	return nil
}
