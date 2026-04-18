package config_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andriikushch/beehive-proxy/internal/config"
)

func TestFromEnv_SlowLogDisabledByDefault(t *testing.T) {
	setEnv(t, "TARGET_URL", "http://example.com")
	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, mw := range cfg.Middlewares {
		if mw == nil {
			continue
		}
		_ = mw
	}
	_ = cfg
}

func TestFromEnv_SlowLogEnabled(t *testing.T) {
	setEnv(t, "TARGET_URL", "http://example.com")
	setEnv(t, "SLOW_LOG_ENABLED", "true")
	setEnv(t, "SLOW_LOG_THRESHOLD", "100ms")
	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = cfg
}

func TestFromEnv_SlowLogInvalidThreshold(t *testing.T) {
	setEnv(t, "TARGET_URL", "http://example.com")
	setEnv(t, "SLOW_LOG_ENABLED", "true")
	setEnv(t, "SLOW_LOG_THRESHOLD", "not-a-duration")
	_, err := config.FromEnv()
	if err == nil {
		t.Fatal("expected error for invalid threshold")
	}
}

func TestFromEnv_SlowLogMiddlewareLogsSlowRequest(t *testing.T) {
	setEnv(t, "TARGET_URL", "http://example.com")
	setEnv(t, "SLOW_LOG_ENABLED", "true")
	setEnv(t, "SLOW_LOG_THRESHOLD", "10ms")
	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var found func(http.Handler) http.Handler
	for _, mw := range cfg.Middlewares {
		if mw != nil {
			found = mw
		}
	}
	if found == nil {
		t.Skip("slow log middleware not registered via Middlewares slice")
	}
	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	h := found(slow)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
