package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andrebq/beehive-proxy/internal/middleware"
)

func captureMethodHandler(t *testing.T) (http.Handler, *string) {
	t.Helper()
	method := new(string)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*method = r.Method
		w.WriteHeader(http.StatusOK)
	})
	return h, method
}

func TestMethodOverride_PassesThroughNonPOST(t *testing.T) {
	h, method := captureMethodHandler(t)
	handler := middleware.NewMethodOverride(h)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-HTTP-Method-Override", "DELETE")
	middleware.NewMethodOverride(h).ServeHTTP(httptest.NewRecorder(), req)
	_ = handler
	if *method != http.MethodGet {
		t.Fatalf("expected GET, got %s", *method)
	}
}

func TestMethodOverride_OverridesViaXHTTPMethodOverride(t *testing.T) {
	h, method := captureMethodHandler(t)
	handler := middleware.NewMethodOverride(h)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-HTTP-Method-Override", "PUT")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if *method != http.MethodPut {
		t.Fatalf("expected PUT, got %s", *method)
	}
}

func TestMethodOverride_OverridesViaQueryParam(t *testing.T) {
	h, method := captureMethodHandler(t)
	handler := middleware.NewMethodOverride(h)
	req := httptest.NewRequest(http.MethodPost, "/?_method=PATCH", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if *method != http.MethodPatch {
		t.Fatalf("expected PATCH, got %s", *method)
	}
}

func TestMethodOverride_IgnoresDisallowedMethod(t *testing.T) {
	h, method := captureMethodHandler(t)
	handler := middleware.NewMethodOverride(h)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-HTTP-Method-Override", "CONNECT")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if *method != http.MethodPost {
		t.Fatalf("expected POST, got %s", *method)
	}
}

func TestMethodOverride_XMethodOverrideFallback(t *testing.T) {
	h, method := captureMethodHandler(t)
	handler := middleware.NewMethodOverride(h)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-Method-Override", "DELETE")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if *method != http.MethodDelete {
		t.Fatalf("expected DELETE, got %s", *method)
	}
}
