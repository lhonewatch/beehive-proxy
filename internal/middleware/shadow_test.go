package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/beehive-proxy/internal/middleware"
)

func okShadowHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestShadow_InvalidURLReturnsError(t *testing.T) {
	_, err := middleware.NewShadow("not-a-url", nil)
	if err == nil {
		t.Fatal("expected error for invalid shadow URL")
	}
}

func TestShadow_PrimaryResponseUnaffected(t *testing.T) {
	shadowSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer shadowSrv.Close()

	mw, err := middleware.NewShadow(shadowSrv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTeapot {
		t.Fatalf("expected 418 got %d", rr.Code)
	}
}

func TestShadow_MirrorReceivesRequest(t *testing.T) {
	var mu sync.Mutex
	var receivedPath string
	received := make(chan struct{}, 1)

	shadowSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		receivedPath = r.URL.Path
		mu.Unlock()
		received <- struct{}{}
		w.WriteHeader(http.StatusOK)
	}))
	defer shadowSrv.Close()

	mw, _ := middleware.NewShadow(shadowSrv.URL, nil)
	handler := mw(http.HandlerFunc(okShadowHandler))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	handler.ServeHTTP(rr, req)

	select {
	case <-received:
	case <-time.After(2 * time.Second):
		t.Fatal("shadow server did not receive request in time")
	}

	mu.Lock()
	defer mu.Unlock()
	if receivedPath != "/hello" {
		t.Fatalf("expected /hello got %s", receivedPath)
	}
}

func TestShadow_SetsXShadowHeader(t *testing.T) {
	headerCh := make(chan string, 1)

	shadowSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerCh <- r.Header.Get("X-Shadow-Request")
		w.WriteHeader(http.StatusOK)
	}))
	defer shadowSrv.Close()

	mw, _ := middleware.NewShadow(shadowSrv.URL, nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/data", nil)
	handler.ServeHTTP(rr, req)

	select {
	case v := <-headerCh:
		if v != "1" {
			t.Fatalf("expected X-Shadow-Request: 1, got %q", v)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out")
	}
}
