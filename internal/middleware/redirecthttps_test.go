package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beehive-proxy/internal/middleware"
)

var httpsOKHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestRedirectHTTPS_RedirectsPlainHTTP(t *testing.T) {
	h := middleware.NewRedirectHTTPS(false)(httpsOKHandler)
	req := httptest.NewRequest(http.MethodGet, "http://example.com/path", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 301, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc != "https://example.com/path" {
		t.Fatalf("unexpected Location: %s", loc)
	}
}

func TestRedirectHTTPS_PassesThroughWhenForwardedProto(t *testing.T) {
	h := middleware.NewRedirectHTTPS(true)(httpsOKHandler)
	req := httptest.NewRequest(http.MethodGet, "http://example.com/secure", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRedirectHTTPS_RedirectsWhenForwardedProtoHTTP(t *testing.T) {
	h := middleware.NewRedirectHTTPS(true)(httpsOKHandler)
	req := httptest.NewRequest(http.MethodGet, "http://example.com/page", nil)
	req.Header.Set("X-Forwarded-Proto", "http")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 301, got %d", rec.Code)
	}
}

func TestRedirectHTTPS_PreservesQueryString(t *testing.T) {
	h := middleware.NewRedirectHTTPS(false)(httpsOKHandler)
	req := httptest.NewRequest(http.MethodGet, "http://example.com/search?q=go", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	loc := rec.Header().Get("Location")
	if loc != "https://example.com/search?q=go" {
		t.Fatalf("unexpected Location: %s", loc)
	}
}
