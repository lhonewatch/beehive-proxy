package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var jwtSecret = []byte("supersecret")

func makeToken(claims map[string]interface{}, secret []byte) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload, _ := json.Marshal(claims)
	p := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(header + "." + p))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return header + "." + p + "." + sig
}

func jwtOKHandler(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }

// newJWTHandler is a test helper that creates a JWT middleware wrapping jwtOKHandler.
func newJWTHandler() http.Handler {
	return NewJWT(JWTOptions{Secret: jwtSecret})(http.HandlerFunc(jwtOKHandler))
}

func TestJWT_AllowsValidToken(t *testing.T) {
	h := newJWTHandler()
	token := makeToken(map[string]interface{}{"sub": "user1"}, jwtSecret)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestJWT_BlocksMissingToken(t *testing.T) {
	h := newJWTHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestJWT_BlocksWrongSignature(t *testing.T) {
	h := newJWTHandler()
	token := makeToken(map[string]interface{}{"sub": "user1"}, []byte("wrongsecret"))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestJWT_ForwardsSubjectHeader(t *testing.T) {
	var gotSub string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSub = r.Header.Get("X-User-ID")
		w.WriteHeader(http.StatusOK)
	})
	h := NewJWT(JWTOptions{Secret: jwtSecret, HeaderKey: "X-User-ID"})(next)
	token := makeToken(map[string]interface{}{"sub": "alice"}, jwtSecret)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if gotSub != "alice" {
		t.Fatalf("expected sub=alice, got %q", gotSub)
	}
}

func TestJWT_BlocksMalformedToken(t *testing.T) {
	h := newJWTHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+strings.Repeat("x", 20))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
