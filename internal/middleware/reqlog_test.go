package middleware_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beehive-proxy/internal/middleware"
)

func echoReqBodyHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func TestRequestBodyLogger_CapturesBody(t *testing.T) {
	var captured []byte
	h := middleware.NewRequestBodyLogger(0, func(b []byte, _ *http.Request) {
		captured = append([]byte{}, b...)
	})(http.HandlerFunc(echoReqBodyHandler))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("hello"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if string(captured) != "hello" {
		t.Fatalf("expected captured body 'hello', got %q", captured)
	}
}

func TestRequestBodyLogger_RestoresBodyForDownstream(t *testing.T) {
	h := middleware.NewRequestBodyLogger(0, nil)(http.HandlerFunc(echoReqBodyHandler))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("world"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Body.String() != "world" {
		t.Fatalf("expected downstream to read 'world', got %q", rec.Body.String())
	}
}

func TestRequestBodyLogger_RespectsMaxSize(t *testing.T) {
	var captured []byte
	h := middleware.NewRequestBodyLogger(3, func(b []byte, _ *http.Request) {
		captured = append([]byte{}, b...)
	})(http.HandlerFunc(echoReqBodyHandler))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("abcdef")))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if string(captured) != "abc" {
		t.Fatalf("expected captured 'abc', got %q", captured)
	}
	// downstream only gets what was buffered (3 bytes)
	if rec.Body.String() != "abc" {
		t.Fatalf("expected downstream 'abc', got %q", rec.Body.String())
	}
}

func TestRequestBodyLogger_NilBodyPassesThrough(t *testing.T) {
	called := false
	h := middleware.NewRequestBodyLogger(0, func(_ []byte, _ *http.Request) {
		called = true
	})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if called {
		t.Fatal("onBody should not be called for nil body")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
