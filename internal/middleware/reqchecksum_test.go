package middleware_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beehive-proxy/internal/middleware"
)

func captureChecksumHandler(t *testing.T, gotHeader *string, gotBody *string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*gotHeader = r.Header.Get("X-Request-Checksum")
		b, _ := io.ReadAll(r.Body)
		*gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	})
}

func TestRequestChecksum_SetsChecksumHeader(t *testing.T) {
	var gotHeader, gotBody string
	h := middleware.NewRequestChecksum(captureChecksumHandler(t, &gotHeader, &gotBody), "")

	body := []byte(`{"hello":"world"}`)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	h.ServeHTTP(rr, req)

	expected := sha256.Sum256(body)
	want := hex.EncodeToString(expected[:])
	if gotHeader != want {
		t.Errorf("expected checksum %s, got %s", want, gotHeader)
	}
}

func TestRequestChecksum_BodyStillReadableDownstream(t *testing.T) {
	var _, gotBody string
	h := middleware.NewRequestChecksum(captureChecksumHandler(t, new(string), &gotBody), "")

	body := []byte("downstream body check")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	h.ServeHTTP(rr, req)

	if gotBody != string(body) {
		t.Errorf("expected body %q downstream, got %q", string(body), gotBody)
	}
}

func TestRequestChecksum_NilBodySetsEmptyHeader(t *testing.T) {
	var gotHeader string
	h := middleware.NewRequestChecksum(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotHeader = r.Header.Get("X-Request-Checksum")
			w.WriteHeader(http.StatusOK)
		}), "",
	)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)

	if gotHeader != "" {
		t.Errorf("expected empty checksum header for nil body, got %q", gotHeader)
	}
}

func TestRequestChecksum_CustomHeaderName(t *testing.T) {
	const customHeader = "X-Body-Hash"
	var gotHeader string
	h := middleware.NewRequestChecksum(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotHeader = r.Header.Get(customHeader)
			w.WriteHeader(http.StatusOK)
		}), customHeader,
	)

	body := []byte("custom header test")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	h.ServeHTTP(rr, req)

	if gotHeader == "" {
		t.Error("expected custom header to be set, got empty string")
	}
	sum := sha256.Sum256(body)
	want := hex.EncodeToString(sum[:])
	if gotHeader != want {
		t.Errorf("expected %s, got %s", want, gotHeader)
	}
}
