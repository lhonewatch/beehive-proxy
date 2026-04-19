package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/yourusername/beehive-proxy/internal/middleware"
)

func okThrottleHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func buildThrottle(rate, burst float64) http.Handler {
	return middleware.NewRequestThrottle(rate, burst, http.HandlerFunc(okThrottleHandler))
}

func TestRequestThrottle_AllowsUnderBurst(t *testing.T) {
	h := buildThrottle(1, 5)
	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, rec.Code)
		}
	}
}

func TestRequestThrottle_BlocksOverBurst(t *testing.T) {
	h := buildThrottle(1, 2)
	send := func() int {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.2:1234"
		h.ServeHTTP(rec, req)
		return rec.Code
	}
	send()
	send()
	if code := send(); code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", code)
	}
}

func TestRequestThrottle_IndependentPerIP(t *testing.T) {
	h := buildThrottle(1, 1)
	for _, ip := range []string{"1.1.1.1:0", "2.2.2.2:0", "3.3.3.3:0"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ip
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("ip %s: expected 200, got %d", ip, rec.Code)
		}
	}
}

func TestRequestThrottle_ReplenishesOverTime(t *testing.T) {
	h := buildThrottle(100, 1) // 100 tokens/sec
	send := func() int {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "5.5.5.5:0"
		h.ServeHTTP(rec, req)
		return rec.Code
	}
	send() // consume burst
	time.Sleep(20 * time.Millisecond) // replenish ~2 tokens
	if code := send(); code != http.StatusOK {
		t.Fatalf("expected 200 after replenish, got %d", code)
	}
}

func TestRequestThrottle_ConcurrentSafe(t *testing.T) {
	h := buildThrottle(1000, 1000)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = "6.6.6.6:0"
			h.ServeHTTP(rec, req)
		}()
	}
	wg.Wait()
}
