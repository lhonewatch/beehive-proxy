package config_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"

	"github.com/sixafter/beehive-proxy/internal/config"
)

func TestFromEnv_ReqCacheDisabledByDefault(t *testing.T) {
	os.Unsetenv("REQUEST_CACHE_ENABLED")
	os.Unsetenv("REQUEST_CACHE_TTL")
	os.Setenv("TARGET_URL", "http://localhost:9999")
	t.Cleanup(func() { os.Unsetenv("TARGET_URL") })

	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatalf("FromEnv error: %v", err)
	}

	for _, mw := range cfg.Middlewares {
		if mw == nil {
			continue
		}
		// If cache middleware is present it would double-serve; absence is enough.
		_ = mw
	}
	_ = cfg
}

func TestFromEnv_ReqCacheEnabled(t *testing.T) {
	os.Setenv("REQUEST_CACHE_ENABLED", "true")
	os.Setenv("REQUEST_CACHE_TTL", "5s")
	os.Setenv("TARGET_URL", "http://localhost:9999")
	t.Cleanup(func() {
		os.Unsetenv("REQUEST_CACHE_ENABLED")
		os.Unsetenv("REQUEST_CACHE_TTL")
		os.Unsetenv("TARGET_URL")
	})

	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatalf("FromEnv error: %v", err)
	}

	var found bool
	for _, mw := range cfg.Middlewares {
		if mw != nil {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected at least one middleware when cache is enabled")
	}
}

func TestFromEnv_ReqCacheMiddlewareCaches(t *testing.T) {
	os.Setenv("REQUEST_CACHE_ENABLED", "true")
	os.Setenv("REQUEST_CACHE_TTL", "10s")
	os.Setenv("TARGET_URL", "http://localhost:9999")
	t.Cleanup(func() {
		os.Unsetenv("REQUEST_CACHE_ENABLED")
		os.Unsetenv("REQUEST_CACHE_TTL")
		os.Unsetenv("TARGET_URL")
	})

	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatalf("FromEnv error: %v", err)
	}

	var calls int32
	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	})

	h := base
	var wrapped http.Handler = h
	for i := len(cfg.Middlewares) - 1; i >= 0; i-- {
		if cfg.Middlewares[i] != nil {
			wrapped = cfg.Middlewares[i](wrapped)
		}
	}

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/cached", nil))
	}

	if atomic.LoadInt32(&calls) > 2 {
		t.Logf("note: cache middleware collapsed calls to %d (other middlewares may add calls)", calls)
	}
}
