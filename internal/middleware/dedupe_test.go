package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourusername/beehive-proxy/internal/middleware"
)

func slowBackendHandler(delay time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello")) //nolint:errcheck
	})
}

func TestDedupe_PassesThroughNonGET(t *testing.T) {
	var calls int32
	h := middleware.NewDedupe()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/data", nil))
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 upstream call, got %d", calls)
	}
}

func TestDedupe_SingleGetPassesThrough(t *testing.T) {
	h := middleware.NewDedupe()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok")) //nolint:errcheck
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/resource", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestDedupe_CollapsesConcurrentRequests(t *testing.T) {
	var calls int32
	h := middleware.NewDedupe()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(40 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("shared")) //nolint:errcheck
	}))

	const n = 5
	var wg sync.WaitGroup
	results := make([]int, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/same", nil))
			results[idx] = rec.Code
		}(i)
	}
	wg.Wait()

	if atomic.LoadInt32(&calls) > 2 {
		t.Fatalf("expected at most 2 upstream calls for %d concurrent reqs, got %d", n, calls)
	}
	for i, code := range results {
		if code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i, code)
		}
	}
}

func TestDedupe_HitHeaderSetOnFollower(t *testing.T) {
	var once sync.Once
	ready := make(chan struct{})
	h := middleware.NewDedupe()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		once.Do(func() { close(ready) })
		time.Sleep(30 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("body")) //nolint:errcheck
	}))

	var wg sync.WaitGroup
	recs := make([]*httptest.ResponseRecorder, 2)
	for i := range recs {
		recs[i] = httptest.NewRecorder()
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			h.ServeHTTP(recs[idx], httptest.NewRequest(http.MethodGet, "/item", nil))
		}(i)
	}
	wg.Wait()

	hitCount := 0
	for _, rec := range recs {
		if rec.Header().Get("X-Dedupe") == "HIT" {
			hitCount++
		}
	}
	if hitCount == 0 {
		t.Error("expected at least one response with X-Dedupe: HIT")
	}
}
