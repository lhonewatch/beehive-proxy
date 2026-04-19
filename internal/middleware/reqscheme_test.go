package middleware_test

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/beehive-proxy/internal/middleware"
)

func captureSchemeHandler(t *testing.T, header string) (http.Handler, func() string) {
	t.Helper()
	var got string
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get(header)
		w.WriteHeader(http.StatusOK)
	})
	return h, func() string { return got }
}

func TestRequestScheme_DefaultsToHTTP(t *testing.T) {
	inner, scheme := captureSchemeHandler(t, "X-Request-Scheme")
	h := middleware.NewRequestScheme("")(inner)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)
	if scheme() != "http" {
		t.Fatalf("expected http, got %s", scheme())
	}
}

func TestRequestScheme_DetectsTLS(t *testing.T) {
	inner, scheme := captureSchemeHandler(t, "X-Request-Scheme")
	h := middleware.NewRequestScheme("")(inner)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.TLS = &tls.ConnectionState{}
	h.ServeHTTP(rec, req)
	if scheme() != "https" {
		t.Fatalf("expected https, got %s", scheme())
	}
}

func TestRequestScheme_PrefersXForwardedProto(t *testing.T) {
	inner, scheme := captureSchemeHandler(t, "X-Request-Scheme")
	h := middleware.NewRequestScheme("")(inner)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.TLS = &tls.ConnectionState{}
	req.Header.Set("X-Forwarded-Proto", "http")
	h.ServeHTTP(rec, req)
	if scheme() != "http" {
		t.Fatalf("expected http from X-Forwarded-Proto, got %s", scheme())
	}
}

func TestRequestScheme_CustomHeader(t *testing.T) {
	inner, scheme := captureSchemeHandler(t, "X-Scheme")
	h := middleware.NewRequestScheme("X-Scheme")(inner)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	h.ServeHTTP(rec, req)
	if scheme() != "https" {
		t.Fatalf("expected https in custom header, got %s", scheme())
	}
}
