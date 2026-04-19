package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/beehive-proxy/internal/middleware"
)

func captureAnnotateHandler(captured *http.Header) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*captured = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	})
}

func buildAnnotateHandler(rules []middleware.AnnotateRule, overwrite bool) (http.Handler, *http.Header) {
	var captured http.Header
	h := middleware.NewRequestAnnotate(rules, overwrite)(captureAnnotateHandler(&captured))
	return h, &captured
}

func TestRequestAnnotate_NoRulesPassesThrough(t *testing.T) {
	h, _ := buildAnnotateHandler(nil, false)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequestAnnotate_InjectsHeader(t *testing.T) {
	rules := []middleware.AnnotateRule{{Header: "X-Env", Value: "production"}}
	h, captured := buildAnnotateHandler(rules, false)
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	if got := (*captured).Get("X-Env"); got != "production" {
		t.Fatalf("expected production, got %q", got)
	}
}

func TestRequestAnnotate_DoesNotOverwriteByDefault(t *testing.T) {
	rules := []middleware.AnnotateRule{{Header: "X-Env", Value: "production"}}
	h, captured := buildAnnotateHandler(rules, false)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Env", "staging")
	h.ServeHTTP(httptest.NewRecorder(), req)
	if got := (*captured).Get("X-Env"); got != "staging" {
		t.Fatalf("expected staging, got %q", got)
	}
}

func TestRequestAnnotate_OverwritesWhenEnabled(t *testing.T) {
	rules := []middleware.AnnotateRule{{Header: "X-Env", Value: "production"}}
	h, captured := buildAnnotateHandler(rules, true)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Env", "staging")
	h.ServeHTTP(httptest.NewRecorder(), req)
	if got := (*captured).Get("X-Env"); got != "production" {
		t.Fatalf("expected production, got %q", got)
	}
}

func TestParseAnnotateRules_ParsesValid(t *testing.T) {
	rules := middleware.ParseAnnotateRules([]string{"X-Region:us-east", "X-Version:v2"})
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
	if rules[0].Header != "X-Region" || rules[0].Value != "us-east" {
		t.Fatalf("unexpected rule: %+v", rules[0])
	}
}

func TestParseAnnotateRules_SkipsMalformed(t *testing.T) {
	rules := middleware.ParseAnnotateRules([]string{"BADENTRY", "X-Good:val"})
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
}
