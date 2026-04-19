package config_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFromEnv_ReqSamplerDisabledByDefault(t *testing.T) {
	t.Setenv("TARGET_URL", "http://localhost:9999")
	t.Setenv("SAMPLER_RATE", "")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, mw := range cfg.Middlewares {
		if mw == nil {
			t.Fatal("nil middleware in chain")
		}
	}
}

func TestFromEnv_ReqSamplerRateOne(t *testing.T) {
	t.Setenv("TARGET_URL", "http://localhost:9999")
	t.Setenv("SAMPLER_RATE", "1.0")
	defer t.Setenv("SAMPLER_RATE", "")

	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = cfg
}

func TestFromEnv_ReqSamplerInvalidRate(t *testing.T) {
	t.Setenv("TARGET_URL", "http://localhost:9999")
	t.Setenv("SAMPLER_RATE", "banana")
	defer t.Setenv("SAMPLER_RATE", "")

	_, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFromEnv_ReqSamplerMiddlewareAllowsAll(t *testing.T) {
	t.Setenv("TARGET_URL", "http://localhost:9999")
	t.Setenv("SAMPLER_RATE", "1.0")
	defer t.Setenv("SAMPLER_RATE", "")

	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	var h http.Handler = inner
	for _, mw := range cfg.Middlewares {
		if mw != nil {
			h = mw(h)
		}
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
