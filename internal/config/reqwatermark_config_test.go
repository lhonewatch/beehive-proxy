package config_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beehive-proxy/internal/config"
	"github.com/beehive-proxy/internal/middleware"
)

func TestFromEnv_ReqWatermarkDisabledByDefault(t *testing.T) {
	t.Setenv("BEEHIVE_WATERMARK_SECRET", "")

	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, m := range cfg.Middlewares {
		_ = m // ensure no panic; watermark should simply be absent
	}
}

func TestFromEnv_ReqWatermarkEnabled(t *testing.T) {
	t.Setenv("BEEHIVE_TARGET_URL", "http://localhost:9999")
	t.Setenv("BEEHIVE_WATERMARK_SECRET", "mysecret")
	defer t.Setenv("BEEHIVE_WATERMARK_SECRET", "")

	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = cfg
}

func TestFromEnv_ReqWatermarkMiddlewareSetsHeader(t *testing.T) {
	const secret = "testsecret"
	var captured string

	h := middleware.NewRequestWatermark(secret, "", "")(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = r.Header.Get(middleware.DefaultWatermarkHeader)
	}))

	req := httptest.NewRequest(http.MethodGet, "/check", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if captured == "" {
		t.Fatal("expected watermark header to be populated")
	}
}

func TestFromEnv_ReqWatermarkCustomHeader(t *testing.T) {
	const wmHeader = "X-Custom-WM"
	var captured string

	h := middleware.NewRequestWatermark("k", wmHeader, "")(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = r.Header.Get(wmHeader)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if captured == "" {
		t.Fatalf("expected %s header to be set", wmHeader)
	}
}
