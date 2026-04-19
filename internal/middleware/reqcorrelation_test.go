package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beehive-proxy/internal/middleware"
)

func okCorrelationHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func buildCorrelationHandler(header string) http.Handler {
	return middleware.NewRequestCorrelation(header, http.HandlerFunc(okCorrelationHandler))
}

func TestRequestCorrelation_GeneratesIDWhenAbsent(t *testing.T) {
	h := buildCorrelationHandler("")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)
	if rec.Header().Get(middleware.DefaultCorrelationHeader) == "" {
		t.Fatal("expected correlation ID in response header")
	}
}

func TestRequestCorrelation_PreservesExistingID(t *testing.T) {
	h := buildCorrelationHandler("")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(middleware.DefaultCorrelationHeader, "my-fixed-id")
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get(middleware.DefaultCorrelationHeader); got != "my-fixed-id" {
		t.Fatalf("expected my-fixed-id, got %s", got)
	}
}

func TestRequestCorrelation_UniquePerRequest(t *testing.T) {
	h := buildCorrelationHandler("")
	ids := make(map[string]struct{})
	for i := 0; i < 20; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		h.ServeHTTP(rec, req)
		id := rec.Header().Get(middleware.DefaultCorrelationHeader)
		if _, dup := ids[id]; dup {
			t.Fatalf("duplicate correlation ID generated: %s", id)
		}
		ids[id] = struct{}{}
	}
}

func TestRequestCorrelation_CustomHeader(t *testing.T) {
	const custom = "X-My-Correlation"
	h := buildCorrelationHandler(custom)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)
	if rec.Header().Get(custom) == "" {
		t.Fatal("expected correlation ID in custom header")
	}
	if rec.Header().Get(middleware.DefaultCorrelationHeader) != "" {
		t.Fatal("default header should not be set when custom header used")
	}
}

func TestRequestCorrelation_PropagatesIDToRequest(t *testing.T) {
	const hdr = middleware.DefaultCorrelationHeader
	var captured string
	inner := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = r.Header.Get(hdr)
	})
	h := middleware.NewRequestCorrelation("", inner)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)
	if captured == "" {
		t.Fatal("expected correlation ID to be set on downstream request")
	}
}
