package middleware_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/beehive-proxy/internal/middleware"
)

func captureWatermarkHandler(t *testing.T, wm, ts *string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		*wm = r.Header.Get(middleware.DefaultWatermarkHeader)
		*ts = r.Header.Get(middleware.DefaultWatermarkTSHeader)
	})
}

func TestRequestWatermark_SetsHeader(t *testing.T) {
	var wm, ts string
	h := middleware.NewRequestWatermark("secret", "", "")(captureWatermarkHandler(t, &wm, &ts))

	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if wm == "" {
		t.Fatal("expected watermark header to be set")
	}
	if ts == "" {
		t.Fatal("expected timestamp header to be set")
	}
}

func TestRequestWatermark_SignatureIsCorrect(t *testing.T) {
	const secret = "topsecret"
	var wm, ts string
	h := middleware.NewRequestWatermark(secret, "", "")(captureWatermarkHandler(t, &wm, &ts))

	req := httptest.NewRequest(http.MethodPost, "/api/v1", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("POST:/api/v1:" + ts))
	want := hex.EncodeToString(mac.Sum(nil))

	if wm != want {
		t.Fatalf("watermark mismatch: got %s want %s", wm, want)
	}
}

func TestRequestWatermark_PreservesExistingTimestamp(t *testing.T) {
	const fixedTS = "2024-01-01T00:00:00Z"
	var _, ts string
	h := middleware.NewRequestWatermark("s", "", "")(captureWatermarkHandler(t, new(string), &ts))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(middleware.DefaultWatermarkTSHeader, fixedTS)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if ts != fixedTS {
		t.Fatalf("expected preserved timestamp %s, got %s", fixedTS, ts)
	}
}

func TestRequestWatermark_CustomHeaders(t *testing.T) {
	const wmH, tsH = "X-My-WM", "X-My-TS"
	var captured string
	h := middleware.NewRequestWatermark("k", wmH, tsH)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = r.Header.Get(wmH)
	}))

	req := httptest.NewRequest(http.MethodDelete, "/del", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if captured == "" {
		t.Fatal("expected custom watermark header to be set")
	}
}

func TestRequestWatermark_UniquePerRequest(t *testing.T) {
	var wm1, wm2 string
	h := middleware.NewRequestWatermark("s", "", "")(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		if wm1 == "" {
			wm1 = r.Header.Get(middleware.DefaultWatermarkHeader)
		} else {
			wm2 = r.Header.Get(middleware.DefaultWatermarkHeader)
		}
	}))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/a", nil))
	time.Sleep(1100 * time.Millisecond)
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/b", nil))

	if wm1 == wm2 {
		t.Fatal("expected different watermarks for different requests")
	}
}
