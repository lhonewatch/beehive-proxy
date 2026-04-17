package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beehive-proxy/internal/middleware"
)

func okMaintenanceHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func TestMaintenanceMode_PassesThroughWhenDisabled(t *testing.T) {
	mm := middleware.NewMaintenanceMode("")
	h := mm.Handler(http.HandlerFunc(okMaintenanceHandler))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestMaintenanceMode_Returns503WhenEnabled(t *testing.T) {
	mm := middleware.NewMaintenanceMode("down for maintenance")
	mm.Enable()
	h := mm.Handler(http.HandlerFunc(okMaintenanceHandler))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}
	if body := rr.Body.String(); body != "down for maintenance" {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestMaintenanceMode_SetsRetryAfterHeader(t *testing.T) {
	mm := middleware.NewMaintenanceMode("")
	mm.Enable()
	h := mm.Handler(http.HandlerFunc(okMaintenanceHandler))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if v := rr.Header().Get("Retry-After"); v != "120" {
		t.Fatalf("expected Retry-After: 120, got %q", v)
	}
}

func TestMaintenanceMode_ReEnablesAfterDisable(t *testing.T) {
	mm := middleware.NewMaintenanceMode("")
	mm.Enable()
	mm.Disable()
	h := mm.Handler(http.HandlerFunc(okMaintenanceHandler))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 after disable, got %d", rr.Code)
	}
}

func TestMaintenanceMode_DefaultBody(t *testing.T) {
	mm := middleware.NewMaintenanceMode("")
	mm.Enable()
	h := mm.Handler(http.HandlerFunc(okMaintenanceHandler))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Body.Len() == 0 {
		t.Fatal("expected non-empty default body")
	}
}
