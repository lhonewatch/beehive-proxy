package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beeehive/beehive-proxy/internal/middleware"
)

func okEchoHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func buildEchoHandler(prefix string, headers []string) http.Handler {
	return middleware.NewRequestEcho(prefix, headers)(http.HandlerFunc(okEchoHandler))
}

func TestRequestEcho_SetsMethodAndPath(t *testing.T) {
	h := buildEchoHandler("X-Echo", nil)
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if got := rr.Header().Get("X-Echo-Method"); got != "GET" {
		t.Fatalf("expected GET, got %q", got)
	}
	if got := rr.Header().Get("X-Echo-Path"); got != "/hello" {
		t.Fatalf("expected /hello, got %q", got)
	}
}

func TestRequestEcho_EchoesRequestedHeader(t *testing.T) {
	h := buildEchoHandler("X-Echo", []string{"X-Trace-Id"})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-Trace-Id", "abc-123")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if got := rr.Header().Get("X-Echo-X-Trace-Id"); got != "abc-123" {
		t.Fatalf("expected abc-123, got %q", got)
	}
}

func TestRequestEcho_SkipsMissingHeader(t *testing.T) {
	h := buildEchoHandler("X-Echo", []string{"X-Missing"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if got := rr.Header().Get("X-Echo-X-Missing"); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestRequestEcho_DefaultPrefixWhenEmpty(t *testing.T) {
	h := buildEchoHandler("", nil)
	req := httptest.NewRequest(http.MethodDelete, "/items/1", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if got := rr.Header().Get("X-Echo-Method"); got != "DELETE" {
		t.Fatalf("expected DELETE, got %q", got)
	}
}

func TestRequestEcho_CustomPrefix(t *testing.T) {
	h := buildEchoHandler("Dbg", []string{"Authorization"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer tok")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if got := rr.Header().Get("Dbg-Authorization"); got != "Bearer tok" {
		t.Fatalf("expected 'Bearer tok', got %q", got)
	}
	if got := rr.Header().Get("Dbg-Method"); got != "GET" {
		t.Fatalf("expected GET, got %q", got)
	}
}
