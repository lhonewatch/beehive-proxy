package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/beehive-proxy/internal/middleware"
)

func okUAHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func buildUABlock(patterns []string, status int) http.Handler {
	opts := middleware.UserAgentBlockOptions{
		Patterns:   patterns,
		StatusCode: status,
	}
	return middleware.NewUserAgentBlock(opts, http.HandlerFunc(okUAHandler))
}

func TestUserAgentBlock_AllowsNonMatchingUA(t *testing.T) {
	h := buildUABlock([]string{"badbot"}, 0)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible)")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestUserAgentBlock_BlocksMatchingUA(t *testing.T) {
	h := buildUABlock([]string{"badbot"}, 0)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "BadBot/1.0")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestUserAgentBlock_CaseInsensitive(t *testing.T) {
	h := buildUABlock([]string{"SCRAPER"}, 0)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "scraper/2.3")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestUserAgentBlock_CustomStatusCode(t *testing.T) {
	h := buildUABlock([]string{"crawler"}, http.StatusTooManyRequests)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "MyCrawler/1.0")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
}

func TestUserAgentBlock_EmptyPatternsAllowsAll(t *testing.T) {
	h := buildUABlock([]string{}, 0)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "anything")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
