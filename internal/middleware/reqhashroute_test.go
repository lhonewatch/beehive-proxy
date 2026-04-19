package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beehive-proxy/internal/middleware"
)

func captureHashBucketHandler(t *testing.T, out *string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*out = r.Header.Get("X-Route-Bucket")
		w.WriteHeader(http.StatusOK)
	})
}

func buildHashRoute(keyHeader string, buckets uint32) func(http.Handler) http.Handler {
	return middleware.NewHashRoute(middleware.HashRouteOptions{
		KeyHeader: keyHeader,
		Buckets:   buckets,
	})
}

func TestHashRoute_SetsBucketHeader(t *testing.T) {
	var got string
	h := buildHashRoute("X-User-ID", 10)(captureHashBucketHandler(t, &got))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-ID", "alice")
	h.ServeHTTP(rec, req)
	if got == "" {
		t.Fatal("expected bucket header to be set")
	}
}

func TestHashRoute_DeterministicForSameKey(t *testing.T) {
	var b1, b2 string
	h := buildHashRoute("X-User-ID", 16)
	for _, out := range []*string{&b1, &b2} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-User-ID", "bob")
		h(captureHashBucketHandler(t, out)).ServeHTTP(rec, req)
	}
	if b1 != b2 {
		t.Fatalf("expected same bucket for same key, got %s and %s", b1, b2)
	}
}

func TestHashRoute_DiffersForDifferentKeys(t *testing.T) {
	results := map[string]bool{}
	h := buildHashRoute("X-User-ID", 100)
	for i := 0; i < 20; i++ {
		var got string
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-User-ID", fmt.Sprintf("user-%d", i))
		h(captureHashBucketHandler(t, &got)).ServeHTTP(rec, req)
		results[got] = true
	}
	if len(results) < 3 {
		t.Fatalf("expected distribution across buckets, got only %d distinct values", len(results))
	}
}

func TestHashRoute_MissingKeyUsesFallback(t *testing.T) {
	var got string
	h := middleware.NewHashRoute(middleware.HashRouteOptions{
		KeyHeader:    "X-User-ID",
		OutputHeader: "X-Route-Bucket",
		Buckets:      10,
		Fallback:     "default",
	})(captureHashBucketHandler(t, &got))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)
	if got != "default" {
		t.Fatalf("expected fallback 'default', got %q", got)
	}
}

func TestHashRoute_MissingKeyNoFallbackSkipsHeader(t *testing.T) {
	var got string
	h := buildHashRoute("X-User-ID", 10)(captureHashBucketHandler(t, &got))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)
	if got != "" {
		t.Fatalf("expected no bucket header, got %q", got)
	}
}
