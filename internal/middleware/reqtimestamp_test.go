package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/beehive-proxy/internal/middleware"
)

func captureTimestampHandler(captured *string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*captured = r.Header.Get(middleware.RequestTimestampHeader)
		w.WriteHeader(http.StatusOK)
	})
}

func TestRequestTimestamp_SetsHeaderWhenAbsent(t *testing.T) {
	var got string
	h := middleware.NewRequestTimestamp(captureTimestampHandler(&got))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got == "" {
		t.Fatal("expected X-Request-Timestamp to be set, got empty string")
	}
	ts, err := time.Parse(time.RFC3339Nano, got)
	if err != nil {
		t.Fatalf("timestamp not RFC3339Nano: %v", err)
	}
	if time.Since(ts) > 5*time.Second {
		t.Errorf("timestamp too old: %v", ts)
	}
}

func TestRequestTimestamp_PreservesExistingHeader(t *testing.T) {
	const existing = "2024-01-01T00:00:00Z"
	var got string
	h := middleware.NewRequestTimestamp(captureTimestampHandler(&got))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(middleware.RequestTimestampHeader, existing)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got != existing {
		t.Errorf("expected existing timestamp %q, got %q", existing, got)
	}
}

func TestRequestTimestamp_UniquePerRequest(t *testing.T) {
	timestamps := make([]string, 3)
	for i := range timestamps {
		h := middleware.NewRequestTimestamp(captureTimestampHandler(&timestamps[i]))
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		h.ServeHTTP(httptest.NewRecorder(), req)
		time.Sleep(2 * time.Millisecond)
	}
	if timestamps[0] == timestamps[1] || timestamps[1] == timestamps[2] {
		t.Error("expected unique timestamps per request")
	}
}
