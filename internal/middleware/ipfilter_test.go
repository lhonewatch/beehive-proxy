package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func okIPHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestIPFilter_AllowlistPermitsMatchingIP(t *testing.T) {
	h := NewIPFilter(Allowlist, []string{"192.168.1.0/24"})(http.HandlerFunc(okIPHandler))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.42:1234"
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestIPFilter_AllowlistBlocksNonMatchingIP(t *testing.T) {
	h := NewIPFilter(Allowlist, []string{"192.168.1.0/24"})(http.HandlerFunc(okIPHandler))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:5678"
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestIPFilter_BlocklistDeniesMatchingIP(t *testing.T) {
	h := NewIPFilter(Blocklist, []string{"10.0.0.1/32"})(http.HandlerFunc(okIPHandler))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:9999"
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestIPFilter_BlocklistAllowsNonMatchingIP(t *testing.T) {
	h := NewIPFilter(Blocklist, []string{"10.0.0.1/32"})(http.HandlerFunc(okIPHandler))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.2:9999"
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestIPFilter_RespectsXForwardedFor(t *testing.T) {
	h := NewIPFilter(Allowlist, []string{"203.0.113.0/24"})(http.HandlerFunc(okIPHandler))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.55, 10.0.0.1")
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestIPFilter_RespectsXRealIP(t *testing.T) {
	h := NewIPFilter(Blocklist, []string{"198.51.100.7/32"})(http.HandlerFunc(okIPHandler))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("X-Real-IP", "198.51.100.7")
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestIPFilter_BareIPNoCIDR(t *testing.T) {
	h := NewIPFilter(Allowlist, []string{"172.16.0.5"})(http.HandlerFunc(okIPHandler))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "172.16.0.5:4321"
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
