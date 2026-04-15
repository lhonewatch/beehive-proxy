package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/beehive-proxy/internal/middleware"
)

func TestHealthCheck_ReturnsOK(t *testing.T) {
	hc := middleware.NewHealthChecker()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	hc.Handler()(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestHealthCheck_ReturnsJSON(t *testing.T) {
	hc := middleware.NewHealthChecker()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	hc.Handler()(rec, req)

	var status middleware.HealthStatus
	if err := json.NewDecoder(rec.Body).Decode(&status); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if status.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", status.Status)
	}
	if status.Uptime == "" {
		t.Error("expected non-empty uptime")
	}
	if status.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestHealthCheck_MiddlewareInterceptsHealthz(t *testing.T) {
	hc := middleware.NewHealthChecker()
	downstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	handler := hc.Middleware(downstream)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected /healthz to return 200, got %d", rec.Code)
	}
}

func TestHealthCheck_MiddlewarePassesThroughOtherPaths(t *testing.T) {
	hc := middleware.NewHealthChecker()
	downstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	handler := hc.Middleware(downstream)
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Errorf("expected downstream 418, got %d", rec.Code)
	}
}

func TestHealthCheck_CountsProxiedRequests(t *testing.T) {
	hc := middleware.NewHealthChecker()
	downstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	handler := hc.Middleware(downstream)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/some/path", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
	// health check should not increment counter
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := hc.RequestCount(); got != 5 {
		t.Errorf("expected 5 proxied requests, got %d", got)
	}
}

func TestHealthCheck_UptimeIncreases(t *testing.T) {
	hc := middleware.NewHealthChecker()
	time.Sleep(10 * time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	hc.Handler()(rec, req)

	var status middleware.HealthStatus
	_ = json.NewDecoder(rec.Body).Decode(&status)
	if status.Uptime == "0s" {
		t.Error("expected uptime > 0s after sleep")
	}
}
