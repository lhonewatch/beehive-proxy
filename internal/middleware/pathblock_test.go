package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beehive-proxy/internal/middleware"
)

func okPathHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func buildPathBlock(opts middleware.PathBlockOptions) http.Handler {
	return middleware.NewPathBlock(opts)(http.HandlerFunc(okPathHandler))
}

func TestPathBlock_AllowsUnblockedPath(t *testing.T) {
	h := buildPathBlock(middleware.PathBlockOptions{ExactPaths: []string{"/blocked"}})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/allowed", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestPathBlock_BlocksExactPath(t *testing.T) {
	h := buildPathBlock(middleware.PathBlockOptions{ExactPaths: []string{"/secret"}})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/secret", nil))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestPathBlock_BlocksPrefixPath(t *testing.T) {
	h := buildPathBlock(middleware.PathBlockOptions{PrefixPaths: []string{"/admin"}})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin/users", nil))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestPathBlock_PrefixDoesNotMatchUnrelated(t *testing.T) {
	h := buildPathBlock(middleware.PathBlockOptions{PrefixPaths: []string{"/admin"}})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/public", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestPathBlock_CustomStatusCode(t *testing.T) {
	h := buildPathBlock(middleware.PathBlockOptions{
		ExactPaths: []string{"/gone"},
		StatusCode: http.StatusNotFound,
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/gone", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
