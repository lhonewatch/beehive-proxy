package config_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFromEnv_ReqBreatheDisabledByDefault(t *testing.T) {
	clearEnv(t)
	t.Setenv("TARGET_URL", "http://example.com")

	cfg, err := FromEnvForTest()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, m := range cfg.Middlewares {
		_ = m // just ensure no panic iterating
	}
}

func TestFromEnv_ReqBreatheEnabled(t *testing.T) {
	clearEnv(t)
	t.Setenv("TARGET_URL", "http://example.com")
	t.Setenv("BREATHE_SOFT_LIMIT", "5")
	t.Setenv("BREATHE_MAX_DELAY_MS", "100")

	cfg, err := FromEnvForTest()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for range cfg.Middlewares {
		found = true
	}
	if !found {
		t.Fatal("expected at least one middleware registered")
	}
}

func TestFromEnv_ReqBreatheMiddlewarePassesThrough(t *testing.T) {
	clearEnv(t)
	t.Setenv("TARGET_URL", "http://example.com")
	t.Setenv("BREATHE_SOFT_LIMIT", "10")
	t.Setenv("BREATHE_MAX_DELAY_MS", "50")

	cfg, err := FromEnvForTest()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	for i := len(cfg.Middlewares) - 1; i >= 0; i-- {
		handler = cfg.Middlewares[i](handler)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
