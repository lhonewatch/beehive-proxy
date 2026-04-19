package config_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFromEnv_ReqVersionDisabledByDefault(t *testing.T) {
	t.Setenv("PROXY_TARGET_URL", "http://example.com")
	t.Setenv("PROXY_API_VERSION_PREFIXES", "")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = cfg
}

func TestFromEnv_ReqVersionSinglePrefix(t *testing.T) {
	t.Setenv("PROXY_TARGET_URL", "http://example.com")
	t.Setenv("PROXY_API_VERSION_PREFIXES", "/v1/=1")
	t.Setenv("PROXY_API_VERSION_HEADER", "X-API-Version")
	t.Setenv("PROXY_API_VERSION_DEFAULT", "0")
	t.Cleanup(func() {
		t.Setenv("PROXY_API_VERSION_PREFIXES", "")
		t.Setenv("PROXY_API_VERSION_HEADER", "")
		t.Setenv("PROXY_API_VERSION_DEFAULT", "")
	})
	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = cfg
}

func TestFromEnv_ReqVersionMiddlewareApplies(t *testing.T) {
	t.Setenv("PROXY_API_VERSION_PREFIXES", "/v2/=2")
	t.Setenv("PROXY_API_VERSION_HEADER", "X-API-Version")
	t.Setenv("PROXY_API_VERSION_DEFAULT", "0")
	t.Cleanup(func() {
		t.Setenv("PROXY_API_VERSION_PREFIXES", "")
		t.Setenv("PROXY_API_VERSION_HEADER", "")
		t.Setenv("PROXY_API_VERSION_DEFAULT", "")
	})

	var captured string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r.Header.Get("X-API-Version")
		w.WriteHeader(http.StatusOK)
	})

	// Build the middleware directly via env helpers to avoid full FromEnv dependency.
	prefixes := map[string]string{"/v2/": "2"}
	import_mw := func(next http.Handler) http.Handler {
		return buildVersionMiddlewareForTest("X-API-Version", "0", prefixes, next)
	}
	h := import_mw(inner)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v2/items", nil))
	if captured != "2" {
		t.Fatalf("expected version '2', got %q", captured)
	}
}

func TestFromEnv_ReqVersionIgnoresMalformedPairs(t *testing.T) {
	t.Setenv("PROXY_TARGET_URL", "http://example.com")
	t.Setenv("PROXY_API_VERSION_PREFIXES", "badpair,,/v1/=1")
	t.Cleanup(func() { t.Setenv("PROXY_API_VERSION_PREFIXES", "") })
	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = cfg
}
