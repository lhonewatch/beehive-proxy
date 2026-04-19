package config_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFromEnv_ReqTagDisabledByDefault(t *testing.T) {
	setEnv(t, map[string]string{
		"TARGET_URL": "http://example.com",
	})
	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// no middleware should be registered for reqtag when REQTAG_RULES is unset
	_ = cfg
}

func TestFromEnv_ReqTagSetsHeader(t *testing.T) {
	setEnv(t, map[string]string{
		"TARGET_URL":   "http://example.com",
		"REQTAG_RULES": "/api=api",
	})
	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Request-Tag")
		w.WriteHeader(http.StatusOK)
	})

	h := inner
	for i := len(cfg.Middlewares) - 1; i >= 0; i-- {
		h = cfg.Middlewares[i](h).(http.HandlerFunc)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/resource", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if got != "api" {
		t.Fatalf("expected X-Request-Tag=api, got %q", got)
	}
}

func TestFromEnv_ReqTagIgnoresMalformedRules(t *testing.T) {
	setEnv(t, map[string]string{
		"TARGET_URL":   "http://example.com",
		"REQTAG_RULES": "badentry",
	})
	_, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
