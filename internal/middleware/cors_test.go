package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var echoHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestCORS_SetsWildcardOrigin(t *testing.T) {
	opts := DefaultCORSOptions()
	handler := NewCORS(opts)(echoHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected *, got %q", got)
	}
}

func TestCORS_PreflightReturns204(t *testing.T) {
	opts := DefaultCORSOptions()
	handler := NewCORS(opts)(echoHandler)

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestCORS_SpecificOriginAllowed(t *testing.T) {
	opts := CORSOptions{
		AllowedOrigins: []string{"https://trusted.com"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Content-Type"},
	}
	handler := NewCORS(opts)(echoHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://trusted.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://trusted.com" {
		t.Errorf("expected https://trusted.com, got %q", got)
	}
}

func TestCORS_UnknownOriginNotSet(t *testing.T) {
	opts := CORSOptions{
		AllowedOrigins: []string{"https://trusted.com"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Content-Type"},
	}
	handler := NewCORS(opts)(echoHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no ACAO header, got %q", got)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestCORS_SetsMaxAge(t *testing.T) {
	opts := DefaultCORSOptions()
	handler := NewCORS(opts)(echoHandler)

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Max-Age"); got != "86400" {
		t.Errorf("expected 86400, got %q", got)
	}
}

func TestCORS_NoOriginHeaderSkipsCORS(t *testing.T) {
	opts := DefaultCORSOptions()
	handler := NewCORS(opts)(echoHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no ACAO header when Origin is absent, got %q", got)
	}
}
