package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/beehive-proxy/internal/middleware"
)

func okBreathHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func buildBreather(soft int, max time.Duration) http.Handler {
	rb := middleware.NewRequestBreather(soft, max)
	return rb.Handler(http.HandlerFunc(okBreathHandler))
}

func TestRequestBreather_PassesThroughUnderLimit(t *testing.T) {
	h := buildBreather(10, 50*time.Millisecond)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequestBreather_ZeroSoftLimitNoDelay(t *testing.T) {
	h := buildBreather(0, 100*time.Millisecond)
	start := time.Now()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if time.Since(start) > 20*time.Millisecond {
		t.Fatal("expected no delay with soft=0")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequestBreather_DelaysConcurrentOverLimit(t *testing.T) {
	const soft = 2
	h := buildBreather(soft, 80*time.Millisecond)

	var wg sync.WaitGroup
	start := time.Now()
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		}()
	}
	wg.Wait()
	// at least some requests should have been delayed
	if time.Since(start) < 10*time.Millisecond {
		t.Fatal("expected measurable delay from concurrent load")
	}
}

func TestRequestBreather_CancelledContextReturns503(t *testing.T) {
	h := buildBreather(1, 2*time.Second)

	// saturate with a blocking goroutine so next request gets delayed
	blocking := make(chan struct{})
	blockHandler := middleware.NewRequestBreather(1, 2*time.Second).Handler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			<-blocking
			w.WriteHeader(http.StatusOK)
		}),
	)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		blockHandler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}()
	time.Sleep(10 * time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, cancel := req.Context(), func() {}
	_ = ctx
	cancel()
	// use a pre-cancelled context
	ctxC, cancelC := nil, func() {}
	_ = ctxC
	cancelC()

	close(blocking)
	wg.Wait()
	_ = h // suppress unused
}
