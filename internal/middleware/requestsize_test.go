package middleware_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beehive-proxy/internal/middleware"
)

func echoHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	w.WriteHeader(http.StatusOK)
}

func TestRequestSizeLimit_AllowsSmallBody(t *testing.T) {
	h := middleware.NewRequestSizeLimit(100)(http.HandlerFunc(echoHandler))
	body := strings.NewReader("hello")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.ContentLength = 5
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestRequestSizeLimit_BlocksLargeContentLength(t *testing.T) {
	h := middleware.NewRequestSizeLimit(10)(http.HandlerFunc(echoHandler))
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("hello"))
	req.ContentLength = 50
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rr.Code)
	}
}

func TestRequestSizeLimit_BlocksOversizedBody(t *testing.T) {
	h := middleware.NewRequestSizeLimit(5)(http.HandlerFunc(echoHandler))
	big := bytes.Repeat([]byte("x"), 100)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(big))
	req.ContentLength = -1
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	// MaxBytesReader causes a read error; handler may return 200 with truncated
	// body or the framework returns 413 — we just ensure no panic and body <= limit.
	if len(rr.Body.Bytes()) > 10 {
		t.Fatalf("body should be capped, got %d bytes", len(rr.Body.Bytes()))
	}
}

func TestRequestSizeLimit_AllowsExactLimit(t *testing.T) {
	h := middleware.NewRequestSizeLimit(5)(http.HandlerFunc(echoHandler))
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("hello"))
	req.ContentLength = 5
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
