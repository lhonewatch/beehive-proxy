package config_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beehive-proxy/internal/config"
)

func TestFromEnv_SecurityHeadersEnabledByDefault(t *testing.T) {
	setEnv(t, map[string]string{"TARGET_URL": "http://example.com"})
	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SecurityHeaders.Middleware() == nil {
		t.Fatal("expected security headers middleware to be non-nil by default")
	}
}

func TestFromEnv_SecurityHeadersDisabled(t *testing.T) {
	setEnv(t, map[string]string{
		"TARGET_URL":                "http://example.com",
		"SECURITY_HEADERS_ENABLED": "false",
	})
	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SecurityHeaders.Middleware() != nil {
		t.Fatal("expected nil middleware when disabled")
	}
}

func TestFromEnv_SecurityHeadersCustomHSTS(t *testing.T) {
	setEnv(t, map[string]string{
		"TARGET_URL":                      "http://example.com",
		"SECURITY_HEADERS_HSTS_MAX_AGE":  "600",
	})
	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SecurityHeaders.HSTSMaxAge != 600 {
		t.Fatalf("expected 600, got %d", cfg.SecurityHeaders.HSTSMaxAge)
	}
}

func TestFromEnv_SecurityHeadersMiddlewareInjects(t *testing.T) {
	setEnv(t, map[string]string{"TARGET_URL": "http://example.com"})
	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	mw := cfg.SecurityHeaders.Middleware()
	handle := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	rec := httptest.NewRecorder()
	handle.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Header().Get("X-Frame-Options") != "SAMEORIGIN" {
		t.Fatal("expected X-Frame-Options header")
	}
}
