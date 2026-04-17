package config_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFromEnv_MaintenanceDisabledByDefault(t *testing.T) {
	setEnv(t, map[string]string{"TARGET_URL": "http://localhost:9090"})
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Maintenance.Enabled {
		t.Fatal("expected maintenance disabled by default")
	}
}

func TestFromEnv_MaintenanceEnabled(t *testing.T) {
	setEnv(t, map[string]string{
		"TARGET_URL":          "http://localhost:9090",
		"MAINTENANCE_ENABLED": "true",
		"MAINTENANCE_BODY":    "brb",
	})
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Maintenance.Enabled {
		t.Fatal("expected maintenance enabled")
	}
	if cfg.Maintenance.Body != "brb" {
		t.Fatalf("unexpected body: %s", cfg.Maintenance.Body)
	}
}

func TestFromEnv_MaintenanceMiddlewareReturns503(t *testing.T) {
	setEnv(t, map[string]string{
		"TARGET_URL":          "http://localhost:9090",
		"MAINTENANCE_ENABLED": "true",
	})
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}

	h := cfg.Maintenance.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}
}
