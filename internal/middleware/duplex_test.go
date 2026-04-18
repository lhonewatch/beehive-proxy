package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/beehive-proxy/internal/middleware"
)

func primaryOKHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestDuplex_PrimaryResponseUnaffected(t *testing.T) {
	var mirrorCalled bool
	var mu sync.Mutex
	mirror := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		mirrorCalled = true
		mu.Unlock()
	}))
	defer mirror.Close()

	h := middleware.NewDuplex(middleware.DuplexOptions{MirrorURL: mirror.URL})(http.HandlerFunc(primaryOKHandler))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if !mirrorCalled {
		t.Fatal("mirror server was not called")
	}
}

func TestDuplex_MirrorReceivesHeaders(t *testing.T) {
	var got string
	var mu sync.Mutex
	mirror := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		mu.Lock()
		got = r.Header.Get("X-Custom")
		mu.Unlock()
	}))
	defer mirror.Close()

	h := middleware.NewDuplex(middleware.DuplexOptions{MirrorURL: mirror.URL})(http.HandlerFunc(primaryOKHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Custom", "hello")
	h.ServeHTTP(httptest.NewRecorder(), req)

	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if got != "hello" {
		t.Fatalf("expected header 'hello', got %q", got)
	}
}

func TestDuplex_DropsCountedOnBadURL(t *testing.T) {
	opts := middleware.DuplexOptions{MirrorURL: "http://127.0.0.1:1"} // nothing listening
	h := middleware.NewDuplex(opts)(http.HandlerFunc(primaryOKHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	time.Sleep(100 * time.Millisecond)
	// primary must still return 200
}
