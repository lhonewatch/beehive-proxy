package config

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFromEnv_JWTDisabledByDefault(t *testing.T) {
	setEnv(t, map[string]string{"TARGET_URL": "http://localhost:9090"})
	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.JWT.Enabled {
		t.Fatal("expected JWT disabled by default")
	}
	if cfg.JWT.JWTMiddleware() != nil {
		t.Fatal("expected nil middleware when disabled")
	}
}

func TestFromEnv_JWTEnabled(t *testing.T) {
	setEnv(t, map[string]string{
		"TARGET_URL":  "http://localhost:9090",
		"JWT_SECRET":  "mysecret",
	})
	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.JWT.Enabled {
		t.Fatal("expected JWT enabled")
	}
	if string(cfg.JWT.Secret) != "mysecret" {
		t.Fatalf("unexpected secret: %s", cfg.JWT.Secret)
	}
	if cfg.JWT.HeaderKey != "X-User-ID" {
		t.Fatalf("unexpected header key: %s", cfg.JWT.HeaderKey)
	}
}

func TestFromEnv_JWTCustomHeader(t *testing.T) {
	setEnv(t, map[string]string{
		"TARGET_URL":           "http://localhost:9090",
		"JWT_SECRET":           "s3cr3t",
		"JWT_SUBJECT_HEADER":   "X-Auth-User",
	})
	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.JWT.HeaderKey != "X-Auth-User" {
		t.Fatalf("expected X-Auth-User, got %s", cfg.JWT.HeaderKey)
	}
}

func TestFromEnv_JWTMiddlewareEnforces(t *testing.T) {
	setEnv(t, map[string]string{
		"TARGET_URL": "http://localhost:9090",
		"JWT_SECRET": "testsecret",
	})
	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mw := cfg.JWT.JWTMiddleware()
	if mw == nil {
		t.Fatal("expected non-nil middleware")
	}
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
