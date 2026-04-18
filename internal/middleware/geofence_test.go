package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beehive-proxy/internal/middleware"
)

func okGeoHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func buildGeoFence(mode string, countries []string) http.Handler {
	opts := middleware.GeoFenceOptions{
		Mode:      mode,
		Countries: countries,
	}
	return middleware.NewGeoFence(opts)(http.HandlerFunc(okGeoHandler))
}

func TestGeoFence_AllowlistPermitsMatchingCountry(t *testing.T) {
	h := buildGeoFence("allowlist", []string{"US", "DE"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Country-Code", "US")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestGeoFence_AllowlistBlocksNonMatchingCountry(t *testing.T) {
	h := buildGeoFence("allowlist", []string{"US"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Country-Code", "CN")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestGeoFence_BlocklistDeniesMatchingCountry(t *testing.T) {
	h := buildGeoFence("blocklist", []string{"RU"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Country-Code", "RU")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestGeoFence_BlocklistAllowsNonMatchingCountry(t *testing.T) {
	h := buildGeoFence("blocklist", []string{"RU"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Country-Code", "FR")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestGeoFence_CaseInsensitiveCountryCode(t *testing.T) {
	h := buildGeoFence("allowlist", []string{"us"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Country-Code", "US")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestGeoFence_MissingHeaderAllowedInBlocklist(t *testing.T) {
	h := buildGeoFence("blocklist", []string{"CN"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
