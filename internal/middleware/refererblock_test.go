package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/radicalbit/beehive-proxy/internal/middleware"
)

func okRefererHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func buildRefererBlock(blocked []string) http.Handler {
	return middleware.NewRefererBlock(blocked)(http.HandlerFunc(okRefererHandler))
}

func TestRefererBlock_AllowsNoReferer(t *testing.T) {
	h := buildRefererBlock([]string{"https://evil.com"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRefererBlock_AllowsNonMatchingReferer(t *testing.T) {
	h := buildRefererBlock([]string{"https://evil.com"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Referer", "https://good.com")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRefererBlock_BlocksExactMatch(t *testing.T) {
	h := buildRefererBlock([]string{"https://evil.com"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Referer", "https://evil.com")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestRefererBlock_BlocksPrefixWildcard(t *testing.T) {
	h := buildRefererBlock([]string{"https://evil.com/*"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Referer", "https://evil.com/path/to/page")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestRefererBlock_CaseInsensitive(t *testing.T) {
	h := buildRefererBlock([]string{"https://Evil.COM"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Referer", "https://evil.com")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestRefererBlock_EmptyListAllowsAll(t *testing.T) {
	h := buildRefererBlock([]string{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Referer", "https://anything.com")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
