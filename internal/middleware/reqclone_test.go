package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/beehive-proxy/internal/middleware"
)

func okCloneHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestRequestClone_PrimaryResponseUnaffected(t *testing.T) {
	cloneSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer cloneSrv.Close()

	h := middleware.NewRequestClone(cloneSrv.URL, nil)(http.HandlerFunc(okCloneHandler))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("hello"))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
}

func TestRequestClone_CloneReceivesBody(t *testing.T) {
	var mu sync.Mutex
	var received string
	done := make(chan struct{})

	cloneSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		received = string(b)
		mu.Unlock()
		close(done)
		w.WriteHeader(http.StatusOK)
	}))
	defer cloneSrv.Close()

	h := middleware.NewRequestClone(cloneSrv.URL, nil)(http.HandlerFunc(okCloneHandler))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("payload"))
	h.ServeHTTP(rr, req)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("clone server never received request")
	}
	mu.Lock()
	defer mu.Unlock()
	if received != "payload" {
		t.Fatalf("expected 'payload' got %q", received)
	}
}

func TestRequestClone_SetsXClonedFromHeader(t *testing.T) {
	var mu sync.Mutex
	var clonePath string
	done := make(chan struct{})

	cloneSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		clonePath = r.Header.Get("X-Cloned-From")
		mu.Unlock()
		close(done)
		w.WriteHeader(http.StatusOK)
	}))
	defer cloneSrv.Close()

	h := middleware.NewRequestClone(cloneSrv.URL, nil)(http.HandlerFunc(okCloneHandler))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	h.ServeHTTP(rr, req)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
	mu.Lock()
	defer mu.Unlock()
	if clonePath != "/api/data" {
		t.Fatalf("expected '/api/data' got %q", clonePath)
	}
}

func TestRequestClone_BadURLDoesNotPanic(t *testing.T) {
	h := middleware.NewRequestClone("http://127.0.0.1:0", nil)(http.HandlerFunc(okCloneHandler))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)
	time.Sleep(50 * time.Millisecond) // let goroutine finish
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
}
