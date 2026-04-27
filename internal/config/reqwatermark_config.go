package config

import (
	"github.com/beehive-proxy/internal/middleware"
)

// reqWatermarkConfigFromEnv reads watermark middleware configuration from
// environment variables and returns a configured middleware or nil.
//
//	BEEHIVE_WATERMARK_SECRET   – HMAC secret (required to enable)
//	BEEHIVE_WATERMARK_HEADER   – request header name (default: X-Watermark)
//	BEEHIVE_WATERMARK_TS_HEADER – timestamp header name (default: X-Watermark-Ts)
func reqWatermarkConfigFromEnv() middlewareEntry {
	secret := envString("BEEHIVE_WATERMARK_SECRET", "")
	if secret == "" {
		return middlewareEntry{}
	}
	header := envString("BEEHIVE_WATERMARK_HEADER", middleware.DefaultWatermarkHeader)
	tsHeader := envString("BEEHIVE_WATERMARK_TS_HEADER", middleware.DefaultWatermarkTSHeader)
	return middlewareEntry{
		Enabled:    true,
		Middleware: middleware.NewRequestWatermark(secret, header, tsHeader),
	}
}

func init() {
	registerMiddlewareFactory(reqWatermarkConfigFromEnv)
}
