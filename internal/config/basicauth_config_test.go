package config_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFromEnv_BasicAuthDisabledByDefault(t *testing.T) {
	setEnv(t, map[string]string{"TARGET_URL": "http://example.com"})
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BasicAuth.Enabled {
		t.Fatal("expected basic auth to be disabled by default")
	}
	if cfg.BasicAuth.BasicAuthMiddleware() != nil {
		t.Fatal("expected nil middleware when disabled")
	}
}

func TestFromEnv_BasicAuthEnabled(t *testing.T) {
	setEnv(t, map[string]string{
		"TARGET_URL":         "http://example.com",
		"BASIC_AUTH_ENABLED": "true",
		"BASIC_AUTH_REALM":   "MyRealm",
		"BASIC_AUTH_USERS":   "alice:pass1,bob:pass2",
	})
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.BasicAuth.Enabled {
		t.Fatal("expected basic auth to be enabled")
	}
	if cfg.BasicAuth.Realm != "MyRealm" {
		t.Fatalf("unexpected realm: %s", cfg.BasicAuth.Realm)
	}
	if cfg.BasicAuth.Credentials["alice"] != "pass1" {
		t.Fatal("expected alice:pass1")
	}
	if cfg.BasicAuth.Credentials["bob"] != "pass2" {
		t.Fatal("expected bob:pass2")
	}
}

func TestFromEnv_BasicAuthMiddlewareEnforces(t *testing.T) {
	setEnv(t, map[string]string{
		"TARGET_URL":         "http://example.com",
		"BASIC_AUTH_ENABLED": "true",
		"BASIC_AUTH_USERS":   "user:hunter2",
	})
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	mw := cfg.BasicAuth.BasicAuthMiddleware()
	if mw == nil {
		t.Fatal("expected non-nil middleware")
	}
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
