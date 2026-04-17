package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beehive-proxy/internal/middleware"
)

func securityOKHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func buildSecurityHandler(opts middleware.SecurityHeadersOptions) http.Handler {
	return middleware.NewSecurityHeaders(opts)(http.HandlerFunc(securityOKHandler))
}

func TestSecurityHeaders_SetsHSTS(t *testing.T) {
	opts := middleware.DefaultSecurityHeadersOptions()
	rec := httptest.NewRecorder()
	buildSecurityHandler(opts).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := rec.Header().Get("Strict-Transport-Security"); got == "" {
		t.Fatal("expected HSTS header")
	}
}

func TestSecurityHeaders_SetsFrameOptions(t *testing.T) {
	opts := middleware.DefaultSecurityHeadersOptions()
	rec := httptest.NewRecorder()
	buildSecurityHandler(opts).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := rec.Header().Get("X-Frame-Options"); got != "SAMEORIGIN" {
		t.Fatalf("expected SAMEORIGIN, got %s", got)
	}
}

func TestSecurityHeaders_SetsNoSniff(t *testing.T) {
	opts := middleware.DefaultSecurityHeadersOptions()
	rec := httptest.NewRecorder()
	buildSecurityHandler(opts).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected nosniff, got %s", got)
	}
}

func TestSecurityHeaders_OmitsHSTSWhenZero(t *testing.T) {
	opts := middleware.DefaultSecurityHeadersOptions()
	opts.HSTSMaxAge = 0
	rec := httptest.NewRecorder()
	buildSecurityHandler(opts).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := rec.Header().Get("Strict-Transport-Security"); got != "" {
		t.Fatalf("expected no HSTS header, got %s", got)
	}
}

func TestSecurityHeaders_SetsPermissionsPolicy(t *testing.T) {
	opts := middleware.DefaultSecurityHeadersOptions()
	opts.PermissionsPolicy = "geolocation=()"
	rec := httptest.NewRecorder()
	buildSecurityHandler(opts).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := rec.Header().Get("Permissions-Policy"); got != "geolocation=()" {
		t.Fatalf("expected geolocation=(), got %s", got)
	}
}
