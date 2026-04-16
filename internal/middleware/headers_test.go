package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beehive-proxy/internal/middleware"
)

func echoHeaderHandler(w http.ResponseWriter, r *http.Request) {
	// Echo a request header back so tests can inspect it.
	if v := r.Header.Get("X-Echo"); v != "" {
		w.Header().Set("X-Echoed", v)
	}
	w.Header().Set("X-Server", "test")
	w.WriteHeader(http.StatusOK)
}

func TestHeaders_InjectsRequestHeader(t *testing.T) {
	mw := middleware.NewHeaders(middleware.HeadersOptions{
		RequestHeaders: map[string]string{"X-Injected": "yes"},
	})
	var got string
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Injected")
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if got != "yes" {
		t.Fatalf("expected X-Injected=yes, got %q", got)
	}
}

func TestHeaders_RemovesRequestHeader(t *testing.T) {
	mw := middleware.NewHeaders(middleware.HeadersOptions{
		RemoveRequestHeaders: []string{"X-Secret"},
	})
	var got string
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Secret")
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Secret", "sensitive")
	h.ServeHTTP(httptest.NewRecorder(), req)
	if got != "" {
		t.Fatalf("expected X-Secret to be removed, got %q", got)
	}
}

func TestHeaders_InjectsResponseHeader(t *testing.T) {
	mw := middleware.NewHeaders(middleware.HeadersOptions{
		ResponseHeaders: map[string]string{"X-Powered-By": "beehive"},
	})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if v := rec.Header().Get("X-Powered-By"); v != "beehive" {
		t.Fatalf("expected X-Powered-By=beehive, got %q", v)
	}
}

func TestHeaders_RemovesResponseHeader(t *testing.T) {
	mw := middleware.NewHeaders(middleware.HeadersOptions{
		RemoveResponseHeaders: []string{"X-Server"},
	})
	h := mw(http.HandlerFunc(echoHeaderHandler))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if v := rec.Header().Get("X-Server"); v != "" {
		t.Fatalf("expected X-Server to be removed, got %q", v)
	}
}

func TestHeaders_DoesNotMutateOriginalRequest(t *testing.T) {
	mw := middleware.NewHeaders(middleware.HeadersOptions{
		RequestHeaders: map[string]string{"X-Added": "1"},
	})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if req.Header.Get("X-Added") != "" {
		t.Fatal("original request was mutated")
	}
}
