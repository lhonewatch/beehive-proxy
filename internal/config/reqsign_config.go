package config

import (
	"github.com/beehive-proxy/internal/middleware"
)

type ReqSignConfig struct {
	Enabled         bool
	Secret          string
	SignatureHeader string
	TimestampHeader string
}

func reqSignConfigFromEnv() ReqSignConfig {
	secret := envString("REQSIGN_SECRET", "")
	return ReqSignConfig{
		Enabled:         secret != "",
		Secret:          secret,
		SignatureHeader: envString("REQSIGN_SIGNATURE_HEADER", "X-Signature"),
		TimestampHeader: envString("REQSIGN_TIMESTAMP_HEADER", "X-Timestamp"),
	}
}

func init() {
	registerMiddleware(func(cfg *Config) optionalMiddleware {
		c := reqSignConfigFromEnv()
		if !c.Enabled {
			return optionalMiddleware{}
		}
		return optionalMiddleware{
			Enabled:    true,
			Middleware: middleware.NewRequestSigning(c.Secret, c.SignatureHeader, c.TimestampHeader),
		}
	})
}
