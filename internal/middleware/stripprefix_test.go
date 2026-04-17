package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beehive-proxy/internal/middleware"
)

func captureStripPathHandler(t *testing.T, got *string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*got = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})
}

func TestStripPrefix_RemovesPrefixFromPath(t *testing.T) {
	var got string
	h := middleware.NewStripPrefix("/api")(captureStripPathHandler(t, &got))
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if got != "/users" {
		t.Fatalf("expected /users, got %s", got)
	}
}

func TestStripPrefix_NoMatchPassesThrough(t *testing.T) {
	var got string
	h := middleware.NewStripPrefix("/api")(captureStripPathHandler(t, &got))
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if got != "/health" {
		t.Fatalf("expected /health, got %s", got)
	}
}

func TestStripPrefix_EmptyPrefixPassesThrough(t *testing.T) {
	var got string
	h := middleware.NewStripPrefix("")(captureStripPathHandler(t, &got))
	req := httptest.NewRequest(http.MethodGet, "/foo/bar", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if got != "/foo/bar" {
		t.Fatalf("expected /foo/bar, got %s", got)
	}
}

func TestStripPrefix_ExactMatchBecomesRoot(t *testing.T) {
	var got string
	h := middleware.NewStripPrefix("/api")(captureStripPathHandler(t, &got))
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if got != "/" {
		t.Fatalf("expected /, got %s", got)
	}
}

func TestStripPrefix_PreservesQueryString(t *testing.T) {
	var got string
	h := middleware.NewStripPrefix("/v1")(captureStripPathHandler(t, &got))
	req := httptest.NewRequest(http.MethodGet, "/v1/items?page=2", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if got != "/items" {
		t.Fatalf("expected /items, got %s", got)
	}
	if req.URL.RawQuery != "page=2" {
		t.Fatalf("query string should be unchanged")
	}
}
