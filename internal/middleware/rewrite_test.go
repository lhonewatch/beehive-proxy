package middleware

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func capturePathHandler(captured *string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*captured = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})
}

func TestRewrite_NoRulesPassesThrough(t *testing.T) {
	var got string
	h := NewRewrite(RewriteOptions{})(capturePathHandler(&got))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if got != "/api/v1/users" {
		t.Errorf("expected /api/v1/users, got %s", got)
	}
}

func TestRewrite_AppliesMatchingRule(t *testing.T) {
	var got string
	opts := RewriteOptions{
		Rules: []RewriteRule{
			{Pattern: regexp.MustCompile(`^/api/v1/(.*)`), Replacement: "/v2/$1"},
		},
	}
	h := NewRewrite(opts)(capturePathHandler(&got))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if got != "/v2/users" {
		t.Errorf("expected /v2/users, got %s", got)
	}
}

func TestRewrite_FirstMatchingRuleWins(t *testing.T) {
	var got string
	opts := RewriteOptions{
		Rules: []RewriteRule{
			{Pattern: regexp.MustCompile(`^/api/(.*)`), Replacement: "/first/$1"},
			{Pattern: regexp.MustCompile(`^/api/(.*)`), Replacement: "/second/$1"},
		},
	}
	h := NewRewrite(opts)(capturePathHandler(&got))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if got != "/first/test" {
		t.Errorf("expected /first/test, got %s", got)
	}
}

func TestRewrite_StripPrefix(t *testing.T) {
	var got string
	opts := RewriteOptions{StripPrefix: "/proxy"}
	h := NewRewrite(opts)(capturePathHandler(&got))

	req := httptest.NewRequest(http.MethodGet, "/proxy/health", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if got != "/health" {
		t.Errorf("expected /health, got %s", got)
	}
}

func TestRewrite_StripPrefixRootFallback(t *testing.T) {
	var got string
	opts := RewriteOptions{StripPrefix: "/proxy"}
	h := NewRewrite(opts)(capturePathHandler(&got))

	req := httptest.NewRequest(http.MethodGet, "/proxy", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if got != "/" {
		t.Errorf("expected /, got %s", got)
	}
}

func TestRewrite_DoesNotMutateOriginalRequest(t *testing.T) {
	var got string
	opts := RewriteOptions{StripPrefix: "/strip"}
	h := NewRewrite(opts)(capturePathHandler(&got))

	req := httptest.NewRequest(http.MethodGet, "/strip/resource", nil)
	originalPath := req.URL.Path
	h.ServeHTTP(httptest.NewRecorder(), req)

	if req.URL.Path != originalPath {
		t.Errorf("original request mutated: got %s, want %s", req.URL.Path, originalPath)
	}
}
