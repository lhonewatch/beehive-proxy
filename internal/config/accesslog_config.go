package config

import (
	"strings"

	"go.uber.org/zap"

	"github.com/patrickward/beehive-proxy/internal/middleware"
)

// AccessLogConfig holds configuration for the access log middleware.
type AccessLogConfig struct {
	Enabled   bool
	SkipPaths []string
}

// accessLogConfigFromEnv reads access-log settings from environment variables:
//
//	ACCESS_LOG_ENABLED   – "true" to enable (default: true)
//	ACCESS_LOG_SKIP      – comma-separated paths to skip (default: "/healthz")
func accessLogConfigFromEnv() AccessLogConfig {
	enabled := envString("ACCESS_LOG_ENABLED", "true") != "false"

	raw := envString("ACCESS_LOG_SKIP", "/healthz")
	var skip []string
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			skip = append(skip, p)
		}
	}

	return AccessLogConfig{
		Enabled:   enabled,
		SkipPaths: skip,
	}
}

// AccessLogMiddleware returns a configured access-log middleware or nil when
// access logging is disabled.
func (c *Config) AccessLogMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	if !c.AccessLog.Enabled {
		return nil
	}
	return middleware.NewAccessLog(middleware.AccessLogOptions{
		Logger:    logger,
		SkipPaths: c.AccessLog.SkipPaths,
	})
}
