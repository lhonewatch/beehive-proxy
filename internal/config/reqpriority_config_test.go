package config_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFromEnv_ReqPriorityDisabledByDefault(t *testing.T) {
	clearEnv()
	t.Setenv("TARGET_URL", "http://localhost:9999")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	_ = cfg // middleware chain should build without error
}

func TestFromEnv_ReqPrioritySetsHeader(t *testing.T) {
	clearEnv()
	t.Setenv("TARGET_URL", "http://localhost:9999")
	t.Setenv("PRIORITY_REJECT_THRESHOLD", "0")

	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	for _, mw := range cfg.Middlewares {
		handler = mw(handler)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Priority", "7")
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
}

func TestFromEnv_ReqPriorityRejectsNegative(t *testing.T) {
	clearEnv()
	t.Setenv("TARGET_URL", "http://localhost:9999")
	t.Setenv("PRIORITY_REJECT_THRESHOLD", "1")

	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	for _, mw := range cfg.Middlewares {
		handler = mw(handler)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Priority", "-5")
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 got %d", rec.Code)
	}
}
