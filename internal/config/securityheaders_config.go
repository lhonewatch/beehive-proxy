package config

import (
	"net/http"

	"github.com/beehive-proxy/internal/middleware"
)

type SecurityHeadersConfig struct {
	Enabled           bool
	HSTSMaxAge        int
	FrameOptions      string
	NoSniff           bool
	XSSProtection     bool
	ReferrerPolicy    string
	PermissionsPolicy string
}

func securityHeadersConfigFromEnv() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		Enabled:        envString("SECURITY_HEADERS_ENABLED", "true") == "true",
		HSTSMaxAge:     envInt("SECURITY_HEADERS_HSTS_MAX_AGE", 31536000),
		FrameOptions:   envString("SECURITY_HEADERS_FRAME_OPTIONS", "SAMEORIGIN"),
		NoSniff:        envString("SECURITY_HEADERS_NO_SNIFF", "true") == "true",
		XSSProtection:  envString("SECURITY_HEADERS_XSS_PROTECTION", "true") == "true",
		ReferrerPolicy: envString("SECURITY_HEADERS_REFERRER_POLICY", "strict-origin-when-cross-origin"),
		PermissionsPolicy: envString("SECURITY_HEADERS_PERMISSIONS_POLICY", ""),
	}
}

func (c SecurityHeadersConfig) Middleware() func(http.Handler) http.Handler {
	if !c.Enabled {
		return nil
	}
	opts := middleware.SecurityHeadersOptions{
		HSTSMaxAge:         c.HSTSMaxAge,
		FrameOptions:       c.FrameOptions,
		ContentTypeNoSniff: c.NoSniff,
		XSSProtection:      c.XSSProtection,
		ReferrerPolicy:     c.ReferrerPolicy,
		PermissionsPolicy:  c.PermissionsPolicy,
	}
	return middleware.NewSecurityHeaders(opts)
}
