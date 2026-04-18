package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func captureFingerprint(captured *string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*captured = r.Header.Get("X-Request-Fingerprint")
		w.WriteHeader(http.StatusOK)
	})
}

func TestRequestFingerprint_SetsHeader(t *testing.T) {
	var fp string
	h := NewRequestFingerprint(nil, captureFingerprint(&fp))
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if fp == "" {
		t.Fatal("expected fingerprint header to be set")
	}
}

func TestRequestFingerprint_DeterministicForSameInput(t *testing.T) {
	var fp1, fp2 string
	h1 := NewRequestFingerprint(nil, captureFingerprint(&fp1))
	h2 := NewRequestFingerprint(nil, captureFingerprint(&fp2))
	req1 := httptest.NewRequest(http.MethodGet, "/same?q=1", nil)
	req2 := httptest.NewRequest(http.MethodGet, "/same?q=1", nil)
	h1.ServeHTTP(httptest.NewRecorder(), req1)
	h2.ServeHTTP(httptest.NewRecorder(), req2)
	if fp1 != fp2 {
		t.Fatalf("expected same fingerprint, got %s and %s", fp1, fp2)
	}
}

func TestRequestFingerprint_DiffersOnDifferentPath(t *testing.T) {
	var fp1, fp2 string
	h1 := NewRequestFingerprint(nil, captureFingerprint(&fp1))
	h2 := NewRequestFingerprint(nil, captureFingerprint(&fp2))
	h1.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/a", nil))
	h2.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/b", nil))
	if fp1 == fp2 {
		t.Fatal("expected different fingerprints for different paths")
	}
}

func TestRequestFingerprint_IncludesExtraHeaders(t *testing.T) {
	var fp1, fp2 string
	h1 := NewRequestFingerprint([]string{"X-Tenant"}, captureFingerprint(&fp1))
	h2 := NewRequestFingerprint([]string{"X-Tenant"}, captureFingerprint(&fp2))
	req1 := httptest.NewRequest(http.MethodGet, "/api", nil)
	req1.Header.Set("X-Tenant", "alpha")
	req2 := httptest.NewRequest(http.MethodGet, "/api", nil)
	req2.Header.Set("X-Tenant", "beta")
	h1.ServeHTTP(httptest.NewRecorder(), req1)
	h2.ServeHTTP(httptest.NewRecorder(), req2)
	if fp1 == fp2 {
		t.Fatal("expected different fingerprints for different tenant headers")
	}
}

func TestRequestFingerprint_LengthIs16Chars(t *testing.T) {
	var fp string
	h := NewRequestFingerprint(nil, captureFingerprint(&fp))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/x", nil))
	if len(fp) != 16 {
		t.Fatalf("expected 16 hex chars, got %d: %s", len(fp), fp)
	}
}
