package middleware_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andrebq/beehive-proxy/internal/middleware"
)

func echoBodyHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func TestBodyTransform_TransformsBody(t *testing.T) {
	upperFn := func(b []byte) ([]byte, error) {
		return []byte(strings.ToUpper(string(b))), nil
	}
	h := middleware.NewBodyTransform(upperFn)(http.HandlerFunc(echoBodyHandler))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("hello"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "HELLO" {
		t.Fatalf("expected HELLO, got %s", got)
	}
}

func TestBodyTransform_NilBodyPassesThrough(t *testing.T) {
	called := false
	h := middleware.NewBodyTransform(func(b []byte) ([]byte, error) {
		called = true
		return b, nil
	})(http.HandlerFunc(echoBodyHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if called {
		t.Fatal("transform should not be called for nil body")
	}
}

func TestBodyTransform_TransformErrorReturns400(t *testing.T) {
	h := middleware.NewBodyTransform(func(b []byte) ([]byte, error) {
		return nil, errors.New("bad")
	})(http.HandlerFunc(echoBodyHandler))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("data"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestBodyTransform_UpdatesContentLength(t *testing.T) {
	h := middleware.NewBodyTransform(func(b []byte) ([]byte, error) {
		return append(b, b...), nil // double the body
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength != 10 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("hello"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
