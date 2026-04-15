package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const compressPayload = "beehive-proxy response body that is long enough to be worth compressing in tests"

func plainHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, _ = io.WriteString(w, compressPayload)
}

func TestCompress_CompressesWhenAccepted(t *testing.T) {
	h := NewCompress(http.HandlerFunc(plainHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected gzip Content-Encoding, got %q", rec.Header().Get("Content-Encoding"))
	}

	gr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gr.Close()

	body, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("failed to read gzip body: %v", err)
	}
	if string(body) != compressPayload {
		t.Fatalf("expected %q, got %q", compressPayload, string(body))
	}
}

func TestCompress_SkipsWhenNotAccepted(t *testing.T) {
	h := NewCompress(http.HandlerFunc(plainHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// no Accept-Encoding header

	h.ServeHTTP(rec, req)

	if enc := rec.Header().Get("Content-Encoding"); enc != "" {
		t.Fatalf("expected no Content-Encoding, got %q", enc)
	}
	if body := rec.Body.String(); body != compressPayload {
		t.Fatalf("expected plain body %q, got %q", compressPayload, body)
	}
}

func TestCompress_PreservesStatusCode(t *testing.T) {
	h := NewCompress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, compressPayload)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestCompress_PoolReuse(t *testing.T) {
	h := NewCompress(http.HandlerFunc(plainHandler))

	for i := 0; i < 10; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		h.ServeHTTP(rec, req)

		if !strings.Contains(rec.Header().Get("Content-Encoding"), "gzip") {
			t.Fatalf("iteration %d: expected gzip encoding", i)
		}
	}
}
