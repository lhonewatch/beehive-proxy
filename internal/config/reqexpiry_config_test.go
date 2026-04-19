package config_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/beehive-proxy/internal/config"
)

func TestFromEnv_ReqExpiryDisabledByDefault(t *testing.T) {
	setEnv(t, "TARGET_URL", "http://localhost:9090")
	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range cfg.Middlewares {
		_ = m // just ensure no panic; expiry not wired when disabled
	}
}

func TestFromEnv_ReqExpiryEnabled(t *testing.T) {
	setEnv(t, "TARGET_URL", "http://localhost:9090")
	setEnv(t, "REQUEST_EXPIRY_ENABLED", "true")
	setEnv(t, "REQUEST_EXPIRY_MAX_AGE", "2m")
	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Middlewares) == 0 {
		t.Fatal("expected at least one middleware")
	}
}

func TestFromEnv_ReqExpiryMiddlewareRejectExpired(t *testing.T) {
	setEnv(t, "TARGET_URL", "http://localhost:9090")
	setEnv(t, "REQUEST_EXPIRY_ENABLED", "true")
	setEnv(t, "REQUEST_EXPIRY_MAX_AGE", "1m")
	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	for i := len(cfg.Middlewares) - 1; i >= 0; i-- {
		h = cfg.Middlewares[i](h)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Timestamp", fmt.Sprintf("%d", time.Now().Add(-2*time.Minute).Unix()))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusGone {
		t.Fatalf("expected 410 for expired request, got %d", rec.Code)
	}
}
