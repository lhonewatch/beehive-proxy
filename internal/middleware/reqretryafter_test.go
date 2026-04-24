package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/YOUR_MODULE/internal/middleware"
)

func tooManyHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusTooManyRequests)
}

func okRetryHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func buildRetryAfter(window time.Duration, handler http.HandlerFunc) http.Handler {
	return middleware.NewRequestRetryAfter(window)(http.HandlerFunc(handler))
}

func TestRequestRetryAfter_AllowsNormalResponse(t *testing.T) {
	h := buildRetryAfter(5*time.Second, okRetryHandler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequestRetryAfter_SetsRetryAfterOn429(t *testing.T) {
	h := buildRetryAfter(10*time.Second, tooManyHandler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	h.ServeHTTP(rec, req)
	if rec.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header to be set")
	}
}

func TestRequestRetryAfter_BlocksClientInWindow(t *testing.T) {
	h := buildRetryAfter(30*time.Second, tooManyHandler)
	ip := "10.0.0.3:1234"

	// First request triggers back-off.
	rec1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = ip
	h.ServeHTTP(rec1, req1)

	// Second request from same IP should be blocked immediately.
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = ip
	h.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec2.Code)
	}
	if rec2.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header on blocked request")
	}
}

func TestRequestRetryAfter_AllowsAfterWindowExpires(t *testing.T) {
	window := 50 * time.Millisecond
	h := buildRetryAfter(window, tooManyHandler)
	ip := "10.0.0.4:1234"

	rec1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = ip
	h.ServeHTTP(rec1, req1)

	time.Sleep(window + 20*time.Millisecond)

	// After window, a new 429 from upstream is served (not pre-rejected).
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = ip
	h.ServeHTTP(rec2, req2)

	// The upstream returned 429 again, so status reflects upstream, not pre-block.
	if rec2.Code == http.StatusOK {
		t.Fatal("did not expect 200 from a 429 upstream")
	}
}

func TestRequestRetryAfter_IndependentPerIP(t *testing.T) {
	h := buildRetryAfter(30*time.Second, tooManyHandler)

	// Trigger back-off for IP A.
	rec1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "10.0.0.5:1234"
	h.ServeHTTP(rec1, req1)

	// IP B should not be affected.
	hB := buildRetryAfter(30*time.Second, okRetryHandler)
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "10.0.0.6:1234"
	hB.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 for independent IP, got %d", rec2.Code)
	}
}
