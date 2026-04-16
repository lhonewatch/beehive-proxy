package middleware_test

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beehive-proxy/internal/middleware"
)

func authOKHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func basicAuthHandler(creds map[string]string) http.Handler {
	opts := middleware.BasicAuthOptions{Credentials: creds, Realm: "TestRealm"}
	return middleware.NewBasicAuth(opts)(http.HandlerFunc(authOKHandler))
}

func setBasicAuth(r *http.Request, user, pass string) {
	encoded := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
	r.Header.Set("Authorization", "Basic "+encoded)
}

func TestBasicAuth_AllowsValidCredentials(t *testing.T) {
	h := basicAuthHandler(map[string]string{"alice": "secret"})
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	setBasicAuth(r, "alice", "secret")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestBasicAuth_BlocksWrongPassword(t *testing.T) {
	h := basicAuthHandler(map[string]string{"alice": "secret"})
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	setBasicAuth(r, "alice", "wrong")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestBasicAuth_BlocksUnknownUser(t *testing.T) {
	h := basicAuthHandler(map[string]string{"alice": "secret"})
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	setBasicAuth(r, "bob", "secret")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestBasicAuth_BlocksMissingHeader(t *testing.T) {
	h := basicAuthHandler(map[string]string{"alice": "secret"})
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	if w.Header().Get("WWW-Authenticate") == "" {
		t.Fatal("expected WWW-Authenticate header")
	}
}

func TestBasicAuth_DefaultRealm(t *testing.T) {
	opts := middleware.BasicAuthOptions{Credentials: map[string]string{"u": "p"}}
	h := middleware.NewBasicAuth(opts)(http.HandlerFunc(authOKHandler))
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Header().Get("WWW-Authenticate") != `Basic realm="Restricted"` {
		t.Fatalf("unexpected realm: %s", w.Header().Get("WWW-Authenticate"))
	}
}
