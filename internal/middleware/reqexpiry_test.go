package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/beehive-proxy/internal/middleware"
)

func okExpiryHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func buildExpiryHandler(maxAge time.Duration) http.Handler {
	return middleware.NewRequestExpiry(maxAge)(http.HandlerFunc(okExpiryHandler))
}

func TestRequestExpiry_AllowsFreshRequest(t *testing.T) {
	h := buildExpiryHandler(5 * time.Minute)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequestExpiry_RejectsExpiredRequest(t *testing.T) {
	h := buildExpiryHandler(1 * time.Minute)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Timestamp", fmt.Sprintf("%d", time.Now().Add(-2*time.Minute).Unix()))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusGone {
		t.Fatalf("expected 410, got %d", rec.Code)
	}
}

func TestRequestExpiry_MissingHeaderReturnsBadRequest(t *testing.T) {
	h := buildExpiryHandler(5 * time.Minute)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRequestExpiry_InvalidTimestampReturnsBadRequest(t *testing.T) {
	h := buildExpiryHandler(5 * time.Minute)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Timestamp", "not-a-number")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
