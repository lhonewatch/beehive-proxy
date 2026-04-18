package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cage1016/beehive-proxy/internal/middleware"
)

func metricsOKHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("metrics"))
}

func buildMetricsAuth(token string) http.Handler {
	return middleware.NewMetricsAuth(token)(http.HandlerFunc(metricsOKHandler))
}

func TestMetricsAuth_AllowsValidToken(t *testing.T) {
	h := buildMetricsAuth("secret")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer secret")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMetricsAuth_BlocksMissingToken(t *testing.T) {
	h := buildMetricsAuth("secret")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestMetricsAuth_BlocksWrongToken(t *testing.T) {
	h := buildMetricsAuth("secret")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestMetricsAuth_SetsWWWAuthenticate(t *testing.T) {
	h := buildMetricsAuth("secret")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	h.ServeHTTP(rec, req)
	if rec.Header().Get("WWW-Authenticate") == "" {
		t.Fatal("expected WWW-Authenticate header")
	}
}

func TestMetricsAuth_NoopWhenEmpty(t *testing.T) {
	h := buildMetricsAuth("")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
