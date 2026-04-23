package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beehive-proxy/internal/middleware"
)

func captureRemoteAddr(addr *string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*addr = r.RemoteAddr
		w.WriteHeader(http.StatusOK)
	})
}

func buildTrustedProxy(cidrs []string) func(http.Handler) http.Handler {
	return middleware.NewTrustedProxy(middleware.TrustedProxyOptions{TrustedCIDRs: cidrs})
}

func TestTrustedProxy_RewritesRemoteAddrWhenTrusted(t *testing.T) {
	var got string
	handler := buildTrustedProxy([]string{"127.0.0.1/32"})(captureRemoteAddr(&got))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:9000"
	req.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if got != "203.0.113.5" {
		t.Fatalf("expected 203.0.113.5, got %s", got)
	}
}

func TestTrustedProxy_SkipsRewriteWhenUntrusted(t *testing.T) {
	var got string
	handler := buildTrustedProxy([]string{"10.0.0.0/8"})(captureRemoteAddr(&got))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.1:1234"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if got != "203.0.113.1:1234" {
		t.Fatalf("expected original RemoteAddr, got %s", got)
	}
}

func TestTrustedProxy_NoXFFPassesThrough(t *testing.T) {
	var got string
	handler := buildTrustedProxy([]string{"127.0.0.0/8"})(captureRemoteAddr(&got))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:5000"
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if got != "127.0.0.1:5000" {
		t.Fatalf("expected unchanged RemoteAddr, got %s", got)
	}
}

func TestTrustedProxy_CIDRRange(t *testing.T) {
	var got string
	handler := buildTrustedProxy([]string{"192.168.1.0/24"})(captureRemoteAddr(&got))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.55:8080"
	req.Header.Set("X-Forwarded-For", "8.8.8.8")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if got != "8.8.8.8" {
		t.Fatalf("expected 8.8.8.8, got %s", got)
	}
}

func TestTrustedProxy_EmptyCIDRsSkipsRewrite(t *testing.T) {
	var got string
	handler := buildTrustedProxy([]string{})(captureRemoteAddr(&got))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:4321"
	req.Header.Set("X-Forwarded-For", "5.6.7.8")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if got != "10.0.0.1:4321" {
		t.Fatalf("expected original RemoteAddr when no CIDRs configured, got %s", got)
	}
}
