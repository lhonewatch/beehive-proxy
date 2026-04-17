package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nicholasgasior/beehive-proxy/internal/middleware"
)

func captureIDHandler(captured *string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*captured = r.Header.Get(middleware.RequestIDHeader)
		w.WriteHeader(http.StatusOK)
	})
}

func TestRequestID_GeneratesIDWhenAbsent(t *testing.T) {
	var got string
	h := middleware.NewRequestID(captureIDHandler(&got))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if got == "" {
		t.Fatal("expected a request ID to be generated")
	}
	if len(got) != 32 {
		t.Fatalf("expected 32-char hex ID, got %q (len %d)", got, len(got))
	}
}

func TestRequestID_ReusesExistingID(t *testing.T) {
	const existing = "abc123"
	var got string
	h := middleware.NewRequestID(captureIDHandler(&got))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(middleware.RequestIDHeader, existing)
	h.ServeHTTP(rec, req)

	if got != existing {
		t.Fatalf("expected %q, got %q", existing, got)
	}
}

func TestRequestID_SetsResponseHeader(t *testing.T) {
	h := middleware.NewRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Header().Get(middleware.RequestIDHeader) == "" {
		t.Fatal("expected X-Request-ID to be set on the response")
	}
}

func TestRequestID_UniquePerRequest(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 50; i++ {
		var got string
		h := middleware.NewRequestID(captureIDHandler(&got))
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		h.ServeHTTP(rec, req)
		if ids[got] {
			t.Fatalf("duplicate request ID generated: %q", got)
		}
		ids[got] = true
	}
}
