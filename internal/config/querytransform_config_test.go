package config_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/beehive-proxy/internal/middleware"
)

func TestFromEnv_QueryTransformDisabledByDefault(t *testing.T) {
	os.Unsetenv("QUERY_TRANSFORM_RULES")
	rules, err := middleware.ParseQueryRewriteRules(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 0 {
		t.Fatalf("expected no rules, got %d", len(rules))
	}
}

func TestFromEnv_QueryTransformSingleRule(t *testing.T) {
	t.Setenv("QUERY_TRANSFORM_RULES", "env:^prod$=production")
	rules, err := middleware.ParseQueryRewriteRules([]string{"env:^prod$=production"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Param != "env" {
		t.Fatalf("unexpected param: %s", rules[0].Param)
	}
}

func TestFromEnv_QueryTransformMiddlewareApplies(t *testing.T) {
	rules, _ := middleware.ParseQueryRewriteRules([]string{"tier:^free$=basic"})
	opts := middleware.QueryTransformOptions{Rules: rules}

	var captured string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r.URL.Query().Get("tier")
		w.WriteHeader(http.StatusOK)
	})

	h := middleware.NewQueryTransform(opts)(inner)
	req := httptest.NewRequest(http.MethodGet, "/?tier=free", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if captured != "basic" {
		t.Fatalf("expected basic, got %s", captured)
	}
}

func TestFromEnv_QueryTransformMultipleRules(t *testing.T) {
	rules, err := middleware.ParseQueryRewriteRules([]string{
		"env:^prod$=production",
		"ver:^1$=v1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
}
