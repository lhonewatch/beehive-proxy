package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beehive-proxy/internal/middleware"
)

func captureVersionHandler(captured *string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*captured = r.Header.Get("X-API-Version")
		w.WriteHeader(http.StatusOK)
	})
}

func buildVersionHandler(defaultVer string, prefixes map[string]string, captured *string) http.Handler {
	return middleware.NewRequestVersion("X-API-Version", defaultVer, prefixes, captureVersionHandler(captured))
}

func TestRequestVersion_SetsDefaultWhenNoMatch(t *testing.T) {
	var got string
	h := buildVersionHandler("0", map[string]string{"/v1/": "1"}, &got)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/unknown/path", nil))
	if got != "0" {
		t.Fatalf("expected default version '0', got %q", got)
	}
}

func TestRequestVersion_MatchesPrefixV1(t *testing.T) {
	var got string
	h := buildVersionHandler("0", map[string]string{"/v1/": "1", "/v2/": "2"}, &got)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/users", nil))
	if got != "1" {
		t.Fatalf("expected version '1', got %q", got)
	}
}

func TestRequestVersion_MatchesPrefixV2(t *testing.T) {
	var got string
	h := buildVersionHandler("0", map[string]string{"/v1/": "1", "/v2/": "2"}, &got)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v2/orders", nil))
	if got != "2" {
		t.Fatalf("expected version '2', got %q", got)
	}
}

func TestRequestVersion_EmptyDefaultAndNoMatch(t *testing.T) {
	var got string
	h := buildVersionHandler("", map[string]string{"/v1/": "1"}, &got)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if got != "" {
		t.Fatalf("expected empty version, got %q", got)
	}
}

func TestRequestVersion_PassesThroughResponse(t *testing.T) {
	var got string
	h := buildVersionHandler("1", map[string]string{}, &got)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/anything", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
