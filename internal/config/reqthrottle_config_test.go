package config_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFromEnv_ReqThrottleDisabledByDefault(t *testing.T) {
	clearEnv()
	t.Setenv("TARGET_URL", "http://localhost:9999")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ThrottleEnabled {
		t.Fatal("expected throttle disabled by default")
	}
}

func TestFromEnv_ReqThrottleEnabled(t *testing.T) {
	clearEnv()
	t.Setenv("TARGET_URL", "http://localhost:9999")
	t.Setenv("THROTTLE_ENABLED", "true")
	t.Setenv("THROTTLE_RATE", "50")
	t.Setenv("THROTTLE_BURST", "100")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.ThrottleEnabled {
		t.Fatal("expected throttle enabled")
	}
	if cfg.ThrottleRate != 50 {
		t.Fatalf("expected rate 50, got %v", cfg.ThrottleRate)
	}
	if cfg.ThrottleBurst != 100 {
		t.Fatalf("expected burst 100, got %v", cfg.ThrottleBurst)
	}
}

func TestFromEnv_ReqThrottleMiddlewareBlocks(t *testing.T) {
	clearEnv()
	t.Setenv("TARGET_URL", "http://localhost:9999")
	t.Setenv("THROTTLE_ENABLED", "true")
	t.Setenv("THROTTLE_RATE", "1")
	t.Setenv("THROTTLE_BURST", "1")

	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	_, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// simulate two rapid requests from same IP via middleware directly
	import_mw := buildThrottleMiddleware(1, 1, ok)

	send := func() int {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "9.9.9.9:0"
		import_mw.ServeHTTP(rec, req)
		return rec.Code
	}
	if code := send(); code != http.StatusOK {
		t.Fatalf("first request: expected 200, got %d", code)
	}
	if code := send(); code != http.StatusTooManyRequests {
		t.Fatalf("second request: expected 429, got %d", code)
	}
}
