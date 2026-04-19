package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andrebq/beehive-proxy/internal/middleware"
)

func captureTagHandler(t *testing.T, got *string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*got = r.Header.Get("X-Request-Tag")
		w.WriteHeader(http.StatusOK)
	})
}

func buildTagHandler(rules []middleware.TagRule) (http.Handler, *string) {
	var tag string
	h := middleware.NewRequestTag(rules, captureTagHandler(nil, &tag))
	return h, &tag
}

func TestRequestTag_NoRulesPassesThrough(t *testing.T) {
	var tag string
	h := middleware.NewRequestTag(nil, captureTagHandler(t, &tag))
	req := httptest.NewRequest(http.MethodGet, "/anything", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if tag != "" {
		t.Fatalf("expected empty tag, got %q", tag)
	}
}

func TestRequestTag_MatchingPrefixSetsTag(t *testing.T) {
	var tag string
	rules := []middleware.TagRule{{Prefix: "/api", Tag: "api"}}
	h := middleware.NewRequestTag(rules, captureTagHandler(t, &tag))
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if tag != "api" {
		t.Fatalf("expected tag 'api', got %q", tag)
	}
}

func TestRequestTag_NonMatchingPrefixSkipped(t *testing.T) {
	var tag string
	rules := []middleware.TagRule{{Prefix: "/api", Tag: "api"}}
	h := middleware.NewRequestTag(rules, captureTagHandler(t, &tag))
	req := httptest.NewRequest(http.MethodGet, "/static/logo.png", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if tag != "" {
		t.Fatalf("expected empty tag, got %q", tag)
	}
}

func TestRequestTag_FirstMatchWins(t *testing.T) {
	var tag string
	rules := []middleware.TagRule{
		{Prefix: "/api", Tag: "api"},
		{Prefix: "/api/v2", Tag: "apiv2"},
	}
	h := middleware.NewRequestTag(rules, captureTagHandler(t, &tag))
	req := httptest.NewRequest(http.MethodGet, "/api/v2/users", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if tag != "api" {
		t.Fatalf("expected first match 'api', got %q", tag)
	}
}

func TestParseTagRules_ParsesValid(t *testing.T) {
	rules := middleware.ParseTagRules("/api=api;/static=static")
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
	if rules[0].Prefix != "/api" || rules[0].Tag != "api" {
		t.Fatalf("unexpected rule[0]: %+v", rules[0])
	}
}

func TestParseTagRules_IgnoresMalformed(t *testing.T) {
	rules := middleware.ParseTagRules("/api=api;badentry;/b=b")
	if len(rules) != 2 {
		t.Fatalf("expected 2 valid rules, got %d", len(rules))
	}
}
