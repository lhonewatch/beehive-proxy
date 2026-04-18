package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/beehive-proxy/internal/middleware"
)

func captureQueryHandler(key string, out *string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*out = r.URL.Query().Get(key)
		w.WriteHeader(http.StatusOK)
	})
}

func TestQueryTransform_NoRulesPassesThrough(t *testing.T) {
	var got string
	h := middleware.NewQueryTransform(middleware.QueryTransformOptions{})(
		captureQueryHandler("q", &got),
	)
	req := httptest.NewRequest(http.MethodGet, "/?q=hello", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if got != "hello" {
		t.Fatalf("expected hello, got %s", got)
	}
}

func TestQueryTransform_RewritesMatchingParam(t *testing.T) {
	var got string
	opts := middleware.QueryTransformOptions{
		Rules: []middleware.QueryRewriteRule{
			{Param: "env", Pattern: regexp.MustCompile(`^prod$`), Replacement: "production"},
		},
	}
	h := middleware.NewQueryTransform(opts)(captureQueryHandler("env", &got))
	req := httptest.NewRequest(http.MethodGet, "/?env=prod", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if got != "production" {
		t.Fatalf("expected production, got %s", got)
	}
}

func TestQueryTransform_SkipsNonMatchingParam(t *testing.T) {
	var got string
	opts := middleware.QueryTransformOptions{
		Rules: []middleware.QueryRewriteRule{
			{Param: "env", Pattern: regexp.MustCompile(`^prod$`), Replacement: "production"},
		},
	}
	h := middleware.NewQueryTransform(opts)(captureQueryHandler("env", &got))
	req := httptest.NewRequest(http.MethodGet, "/?env=staging", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if got != "staging" {
		t.Fatalf("expected staging, got %s", got)
	}
}

func TestParseQueryRewriteRules_ParsesValid(t *testing.T) {
	rules, err := middleware.ParseQueryRewriteRules([]string{"version:^v1$=v2"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 1 || rules[0].Param != "version" || rules[0].Replacement != "v2" {
		t.Fatalf("unexpected rules: %+v", rules)
	}
}

func TestParseQueryRewriteRules_InvalidRegex(t *testing.T) {
	_, err := middleware.ParseQueryRewriteRules([]string{"p:[invalid=x"})
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}
