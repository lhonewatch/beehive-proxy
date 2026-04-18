package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/yourorg/beehive-proxy/internal/middleware"
)

func okCounterHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestRequestCounter_StartsAtZero(t *testing.T) {
	rc := middleware.NewRequestCounter(http.HandlerFunc(okCounterHandler))
	if rc.Total() != 0 {
		t.Fatalf("expected 0, got %d", rc.Total())
	}
}

func TestRequestCounter_IncrementsOnEachRequest(t *testing.T) {
	rc := middleware.NewRequestCounter(http.HandlerFunc(okCounterHandler))
	for i := 1; i <= 5; i++ {
		rec := httptest.NewRecorder()
		rc.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rc.Total() != int64(i) {
			t.Fatalf("after %d requests expected total %d, got %d", i, i, rc.Total())
		}
	}
}

func TestRequestCounter_ConcurrentSafe(t *testing.T) {
	const n = 100
	rc := middleware.NewRequestCounter(http.HandlerFunc(okCounterHandler))
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			rc.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		}()
	}
	wg.Wait()
	if rc.Total() != n {
		t.Fatalf("expected %d, got %d", n, rc.Total())
	}
}

func TestRequestCounter_DownstreamStillReceivesRequest(t *testing.T) {
	var called bool
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	rc := middleware.NewRequestCounter(h)
	rec := httptest.NewRecorder()
	rc.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if !called {
		t.Fatal("downstream handler was not called")
	}
}
