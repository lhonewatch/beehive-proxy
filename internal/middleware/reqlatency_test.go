package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/beehive-proxy/internal/middleware"
)

func okLatencyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func slowLatencyHandler(d time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(d)
		w.WriteHeader(http.StatusOK)
	}
}

func buildLatencyHandler(threshold time.Duration, h http.Handler) http.Handler {
	return middleware.NewRequestLatency(threshold)(h)
}

func TestRequestLatency_SetsHeader(t *testing.T) {
	h := buildLatencyHandler(0, http.HandlerFunc(okLatencyHandler))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	v := rec.Header().Get("X-Request-Latency")
	if v == "" {
		t.Fatal("expected X-Request-Latency header to be set")
	}
	if !strings.HasSuffix(v, "ms") {
		t.Fatalf("expected ms suffix, got %q", v)
	}
}

func TestRequestLatency_ValueIsNumeric(t *testing.T) {
	h := buildLatencyHandler(0, http.HandlerFunc(okLatencyHandler))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	v := rec.Header().Get("X-Request-Latency")
	numPart := strings.TrimSuffix(v, "ms")
	if _, err := strconv.ParseInt(numPart, 10, 64); err != nil {
		t.Fatalf("non-numeric latency value %q: %v", v, err)
	}
}

func TestRequestLatency_SlowHeaderAbsentWhenFast(t *testing.T) {
	h := buildLatencyHandler(10*time.Second, http.HandlerFunc(okLatencyHandler))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Header().Get("X-Request-Latency-Slow") != "" {
		t.Fatal("expected slow header to be absent for fast request")
	}
}

func TestRequestLatency_SlowHeaderSetWhenSlow(t *testing.T) {
	h := buildLatencyHandler(10*time.Millisecond, slowLatencyHandler(30*time.Millisecond))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Header().Get("X-Request-Latency-Slow") != "true" {
		t.Fatal("expected X-Request-Latency-Slow: true")
	}
}

func TestRequestLatency_ZeroThresholdNeverSetsSlowHeader(t *testing.T) {
	h := buildLatencyHandler(0, slowLatencyHandler(20*time.Millisecond))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Header().Get("X-Request-Latency-Slow") != "" {
		t.Fatal("slow header should not be set when threshold is 0")
	}
}
