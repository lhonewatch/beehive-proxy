package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beehive-proxy/internal/middleware"
)

func writeLargeBodyHandler(body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	})
}

func TestResponseSizeLimit_AllowsSmallResponse(t *testing.T) {
	h := middleware.NewResponseSizeLimit(100)(writeLargeBodyHandler("hello"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Body.String() != "hello" {
		t.Fatalf("expected 'hello', got %q", rec.Body.String())
	}
}

func TestResponseSizeLimit_TruncatesLargeResponse(t *testing.T) {
	body := strings.Repeat("a", 200)
	h := middleware.NewResponseSizeLimit(50)(writeLargeBodyHandler(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Body.Len() > 50 {
		t.Fatalf("expected body <= 50 bytes, got %d", rec.Body.Len())
	}
}

func TestResponseSizeLimit_ZeroDisablesLimit(t *testing.T) {
	body := strings.Repeat("b", 500)
	h := middleware.NewResponseSizeLimit(0)(writeLargeBodyHandler(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Body.Len() != 500 {
		t.Fatalf("expected 500 bytes, got %d", rec.Body.Len())
	}
}

func TestResponseSizeLimit_ExactLimit(t *testing.T) {
	body := strings.Repeat("c", 64)
	h := middleware.NewResponseSizeLimit(64)(writeLargeBodyHandler(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Body.Len() != 64 {
		t.Fatalf("expected 64 bytes, got %d", rec.Body.Len())
	}
}
