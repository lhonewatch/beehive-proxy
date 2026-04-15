package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/beehive-proxy/internal/middleware"
)

func writingHandler(body string, status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		w.Write([]byte(body)) //nolint:errcheck
	})
}

func TestCacheMiddleware_CachesGETResponse(t *testing.T) {
	var calls int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello")) //nolint:errcheck
	})

	cache := middleware.NewResponseCache(10 * time.Second)
	mw := middleware.NewCacheMiddleware(cache)(handler)

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/foo", nil))
	}

	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("expected backend called once, got %d", calls)
	}
}

func TestCacheMiddleware_SetsHitHeader(t *testing.T) {
	handler := writingHandler("data", http.StatusOK)
	cache := middleware.NewResponseCache(10 * time.Second)
	mw := middleware.NewCacheMiddleware(cache)(handler)

	// prime the cache
	mw.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/bar", nil))

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/bar", nil))

	if rec.Header().Get("X-Cache") != "HIT" {
		t.Errorf("expected X-Cache: HIT, got %q", rec.Header().Get("X-Cache"))
	}
}

func TestCacheMiddleware_SkipsNonGET(t *testing.T) {
	var calls int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	})

	cache := middleware.NewResponseCache(10 * time.Second)
	mw := middleware.NewCacheMiddleware(cache)(handler)

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/baz", nil))
	}

	if atomic.LoadInt32(&calls) != 3 {
		t.Errorf("expected backend called 3 times for POST, got %d", calls)
	}
}

func TestCacheMiddleware_ExpiresTTL(t *testing.T) {
	var calls int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	})

	cache := middleware.NewResponseCache(50 * time.Millisecond)
	mw := middleware.NewCacheMiddleware(cache)(handler)

	mw.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ttl", nil))
	time.Sleep(100 * time.Millisecond)
	mw.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ttl", nil))

	if atomic.LoadInt32(&calls) != 2 {
		t.Errorf("expected backend called twice after TTL expiry, got %d", calls)
	}
}
