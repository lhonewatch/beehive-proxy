package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/beehive-proxy/internal/middleware"
)

func okBurstHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func buildBurst(limit int64, window time.Duration) http.Handler {
	return middleware.NewRequestBurst(limit, window)(http.HandlerFunc(okBurstHandler))
}

func TestRequestBurst_AllowsUnderLimit(t *testing.T) {
	h := buildBurst(5, time.Second)
	for i := 0; i < 5; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rr.Code)
		}
	}
}

func TestRequestBurst_BlocksOverLimit(t *testing.T) {
	h := buildBurst(3, time.Second)
	var last int
	for i := 0; i < 5; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.2:1234"
		h.ServeHTTP(rr, req)
		last = rr.Code
	}
	if last != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", last)
	}
}

func TestRequestBurst_IndependentPerIP(t *testing.T) {
	h := buildBurst(2, time.Second)
	for _, ip := range []string{"10.0.0.3", "10.0.0.4"} {
		for i := 0; i < 2; i++ {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = ip + ":9999"
			h.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("ip %s request %d: expected 200, got %d", ip, i+1, rr.Code)
			}
		}
	}
}

func TestRequestBurst_ZeroLimitDisables(t *testing.T) {
	h := buildBurst(0, time.Second)
	for i := 0; i < 20; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.5:1234"
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 when limit=0, got %d", rr.Code)
		}
	}
}

func TestRequestBurst_ConcurrentSafe(t *testing.T) {
	h := buildBurst(100, time.Second)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = "10.0.0.6:1234"
			h.ServeHTTP(rr, req)
		}()
	}
	wg.Wait()
}
