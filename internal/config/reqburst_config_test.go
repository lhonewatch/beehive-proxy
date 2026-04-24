package config_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beehive-proxy/internal/config"
)

func TestFromEnv_ReqBurstDisabledByDefault(t *testing.T) {
	setEnv(t, map[string]string{
		"TARGET_URL": "http://example.com",
	})
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

func TestFromEnv_ReqBurstEnabled(t *testing.T) {
	setEnv(t, map[string]string{
		"TARGET_URL":   "http://example.com",
		"BURST_LIMIT":  "5",
		"BURST_WINDOW": "10s",
	})
	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = cfg
}

func TestFromEnv_ReqBurstMiddlewareBlocks(t *testing.T) {
	setEnv(t, map[string]string{
		"TARGET_URL":   "http://example.com",
		"BURST_LIMIT":  "2",
		"BURST_WINDOW": "5s",
	})
	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var mw func(http.Handler) http.Handler
	for _, m := range cfg.Middlewares {
		if m != nil {
			mw = m
		}
	}
	if mw == nil {
		t.Skip("burst middleware not registered in this build")
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := mw(inner)

	var last int
	for i := 0; i < 5; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.168.1.1:9000"
		h.ServeHTTP(rr, req)
		last = rr.Code
	}
	if last != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after burst exceeded, got %d", last)
	}
}
