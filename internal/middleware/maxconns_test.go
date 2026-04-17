package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/beehive-proxy/internal/middleware"
)

func blockedHandler(ready, release chan struct{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(ready)
		<-release
		w.WriteHeader(http.StatusOK)
	})
}

func TestMaxConns_AllowsUnderLimit(t *testing.T) {
	h := middleware.NewMaxConns(2, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMaxConns_BlocksOverLimit(t *testing.T) {
	ready := make(chan struct{})
	release := make(chan struct{})

	h := middleware.NewMaxConns(1, blockedHandler(ready, release))

	// occupy the single slot
	go func() {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}()

	// wait until slot is occupied
	select {
	case <-ready:
	case <-time.After(time.Second):
		t.Fatal("handler never started")
	}

	// second request should be rejected
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}

	close(release)
}

func TestMaxConns_ZeroDisablesLimit(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	h := middleware.NewMaxConns(0, inner)
	if h != inner {
		t.Fatal("expected unwrapped handler when n=0")
	}
}

func TestMaxConns_ConcurrentUnderLimit(t *testing.T) {
	const limit = 5
	var active int64
	var mu sync.Mutex

	h := middleware.NewMaxConns(limit, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		active++
		if active > limit {
			t.Errorf("active %d exceeded limit %d", active, limit)
		}
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		mu.Lock()
		active--
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))

	var wg sync.WaitGroup
	for i := 0; i < limit; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		}()
	}
	wg.Wait()
}
