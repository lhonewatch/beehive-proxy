package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sixafter/beehive-proxy/internal/middleware"
)

func buildCacheHandler(ttl time.Duration, next http.Handler) http.Handler {
	rc := middleware.NewRequestCache(ttl)
	return rc.Handler(next)
}

func TestRequestCache_CachesGETResponse(t *testing.T) {
	var calls int32
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	})

	h := buildCacheHandler(5*time.Second, backend)

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/foo", nil)
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	}

	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected backend called once, got %d", calls)
	}
}

func TestRequestCache_SetsHitHeader(t *testing.T) {
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := buildCacheHandler(5*time.Second, backend)

	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/bar", nil))
		if i == 1 && rec.Header().Get("X-Request-Cache") != "HIT" {
			t.Fatal("expected X-Request-Cache: HIT on second request")
		}
	}
}

func TestRequestCache_SkipsNonGET(t *testing.T) {
	var calls int32
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	})
	h := buildCacheHandler(5*time.Second, backend)

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/baz", nil))
	}

	if atomic.LoadInt32(&calls) != 3 {
		t.Fatalf("expected 3 backend calls for POST, got %d", calls)
	}
}

func TestRequestCache_ExpiresTTL(t *testing.T) {
	var calls int32
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	})

	rc := middleware.NewRequestCache(10 * time.Millisecond)
	h := rc.Handler(backend)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ttl", nil))

	time.Sleep(20 * time.Millisecond)

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ttl", nil))

	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 backend calls after TTL expiry, got %d", calls)
	}
}

func TestRequestCache_DoesNotCacheNon2xx(t *testing.T) {
	var calls int32
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
	})
	h := buildCacheHandler(5*time.Second, backend)

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/err", nil))
	}

	if atomic.LoadInt32(&calls) != 3 {
		t.Fatalf("expected 3 backend calls for non-2xx, got %d", calls)
	}
}
