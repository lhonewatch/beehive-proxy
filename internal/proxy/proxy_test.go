package proxy_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/beehive-proxy/internal/proxy"
	"github.com/beehive-proxy/internal/tracing"
)

func newTestBackend(statusCode int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(body))
	}))
}

func TestHandler_ProxiesRequest(t *testing.T) {
	backend := newTestBackend(http.StatusOK, "hello")
	defer backend.Close()

	target, _ := url.Parse(backend.URL)
	h := proxy.NewHandler(target, tracing.DefaultTracerFunc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_PropagatesTraceID(t *testing.T) {
	var capturedTraceID string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTraceID = r.Header.Get(tracing.TraceIDHeader)
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	target, _ := url.Parse(backend.URL)
	h := proxy.NewHandler(target, tracing.DefaultTracerFunc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(tracing.TraceIDHeader, "test-trace-123")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if capturedTraceID != "test-trace-123" {
		t.Fatalf("expected trace id 'test-trace-123', got %q", capturedTraceID)
	}
}
