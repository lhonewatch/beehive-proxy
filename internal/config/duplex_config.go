package config

import (
	"net/http"
	"time"

	"github.com/beehive-proxy/internal/middleware"
)

type DuplexConfig struct {
	Enabled   bool
	MirrorURL string
	Timeout   time.Duration
}

func duplexConfigFromEnv() DuplexConfig {
	url := envString("DUPLEX_MIRROR_URL", "")
	return DuplexConfig{
		Enabled:   url != "",
		MirrorURL: url,
		Timeout:   envDuration("DUPLEX_TIMEOUT", 2*time.Second),
	}
}

func init() {
	registerMiddleware(func(cfg *Config) func(http.Handler) http.Handler {
		dc := duplexConfigFromEnv()
		if !dc.Enabled {
			return nil
		}
		client := &http.Client{Timeout: dc.Timeout}
		return middleware.NewDuplex(middleware.DuplexOptions{
			MirrorURL: dc.MirrorURL,
			Client:    client,
		})
	})
}
