package middleware_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beehive-proxy/internal/middleware"
)

func signExpected(secret, method, path, ts string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(method + "\n" + path + "\n" + ts))
	return hex.EncodeToString(mac.Sum(nil))
}

func captureSignHandler(sigHeader, tsHeader *string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*sigHeader = r.Header.Get("X-Signature")
		*tsHeader = r.Header.Get("X-Timestamp")
		w.WriteHeader(http.StatusOK)
	})
}

func TestRequestSigning_SetsSigHeader(t *testing.T) {
	var sig, ts string
	h := middleware.NewRequestSigning("secret", "", "")(captureSignHandler(&sig, &ts))
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if sig == "" {
		t.Fatal("expected X-Signature to be set")
	}
	if ts == "" {
		t.Fatal("expected X-Timestamp to be set")
	}
}

func TestRequestSigning_SignatureIsCorrect(t *testing.T) {
	var sig, ts string
	h := middleware.NewRequestSigning("mysecret", "", "")(captureSignHandler(&sig, &ts))
	req := httptest.NewRequest(http.MethodPost, "/api/data", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	expected := signExpected("mysecret", http.MethodPost, "/api/data", ts)
	if sig != expected {
		t.Errorf("signature mismatch: got %s, want %s", sig, expected)
	}
}

func TestRequestSigning_ReusesExistingTimestamp(t *testing.T) {
	var sig, ts string
	h := middleware.NewRequestSigning("secret", "", "")(captureSignHandler(&sig, &ts))
	req := httptest.NewRequest(http.MethodGet, "/path", nil)
	req.Header.Set("X-Timestamp", "2024-01-01T00:00:00Z")
	h.ServeHTTP(httptest.NewRecorder(), req)
	if ts != "2024-01-01T00:00:00Z" {
		t.Errorf("expected existing timestamp to be preserved, got %s", ts)
	}
}

func TestRequestSigning_CustomHeaders(t *testing.T) {
	var captured string
	h := middleware.NewRequestSigning("s", "X-Custom-Sig", "X-Custom-Ts")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r.Header.Get("X-Custom-Sig")
		w.WriteHeader(200)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if captured == "" {
		t.Fatal("expected X-Custom-Sig to be set")
	}
}
