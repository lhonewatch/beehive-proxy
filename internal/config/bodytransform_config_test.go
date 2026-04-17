package config_test

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andrebq/beehive-proxy/internal/config"
)

func TestFromEnv_BodyTransformDisabledByDefault(t *testing.T) {
	t.Setenv("BODY_TRANSFORM_MODE", "")
	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BodyTransform.Enabled {
		t.Fatal("expected body transform disabled by default")
	}
}

func TestFromEnv_BodyTransformUppercase(t *testing.T) {
	t.Setenv("BODY_TRANSFORM_MODE", "uppercase")
	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.BodyTransform.Enabled {
		t.Fatal("expected body transform enabled")
	}
	if cfg.BodyTransform.Mode != "uppercase" {
		t.Fatalf("unexpected mode: %s", cfg.BodyTransform.Mode)
	}
}

func TestFromEnv_BodyTransformMiddlewareAppliesUppercase(t *testing.T) {
	t.Setenv("BODY_TRANSFORM_MODE", "uppercase")
	cfg, _ := config.FromEnv()
	mw := cfg.BodyTransform.Middleware()
	if mw == nil {
		t.Fatal("expected non-nil middleware")
	}

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_, _ = w.Write(body)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("hello"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Body.String(); got != "HELLO" {
		t.Fatalf("expected HELLO got %s", got)
	}
}

func TestFromEnv_BodyTransformBase64RoundTrip(t *testing.T) {
	original := []byte("beehive")
	encoded := base64.StdEncoding.EncodeToString(original)

	t.Setenv("BODY_TRANSFORM_MODE", "base64decode")
	cfg, _ := config.FromEnv()
	mw := cfg.BodyTransform.Middleware()

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_, _ = w.Write(body)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(encoded))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Body.Bytes(); !bytes.Equal(got, original) {
		t.Fatalf("expected %s got %s", original, got)
	}
}
