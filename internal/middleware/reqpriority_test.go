package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/beehive-proxy/internal/middleware"
)

func okPriorityHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestRequestPriority_SetsHeader(t *testing.T) {
	h := middleware.NewRequestPriority(http.HandlerFunc(okPriorityHandler), 0, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Priority", "5")
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("X-Request-Priority"); got != "5" {
		t.Fatalf("expected 5 got %s", got)
	}
}

func TestRequestPriority_DefaultsToZero(t *testing.T) {
	h := middleware.NewRequestPriority(http.HandlerFunc(okPriorityHandler), 0, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("X-Request-Priority"); got != "0" {
		t.Fatalf("expected 0 got %s", got)
	}
}

func TestRequestPriority_RejectsNegativeWhenThresholdSet(t *testing.T) {
	h := middleware.NewRequestPriority(http.HandlerFunc(okPriorityHandler), 1, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Priority", "-1")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 got %d", rec.Code)
	}
}

func TestRequestPriority_AllowsPositiveEvenWithThreshold(t *testing.T) {
	h := middleware.NewRequestPriority(http.HandlerFunc(okPriorityHandler), 1, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Priority", "3")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
}

func TestRequestPriority_CustomFn(t *testing.T) {
	fn := func(_ *http.Request) int { return 99 }
	h := middleware.NewRequestPriority(http.HandlerFunc(okPriorityHandler), 0, fn)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("X-Request-Priority"); got != "99" {
		t.Fatalf("expected 99 got %s", got)
	}
}
