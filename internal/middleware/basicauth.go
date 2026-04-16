package middleware

import (
	"crypto/subtle"
	"net/http"
)

// BasicAuthOptions configures the BasicAuth middleware.
type BasicAuthOptions struct {
	// Credentials maps username -> password.
	Credentials map[string]string
	Realm       string
}

// NewBasicAuth returns a middleware that enforces HTTP Basic Authentication.
// Requests without valid credentials receive a 401 response.
func NewBasicAuth(opts BasicAuthOptions) func(http.Handler) http.Handler {
	realm := opts.Realm
	if realm == "" {
		realm = "Restricted"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok {
				challenge(w, realm)
				return
			}
			expected, found := opts.Credentials[user]
			if !found || subtle.ConstantTimeCompare([]byte(pass), []byte(expected)) != 1 {
				challenge(w, realm)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func challenge(w http.ResponseWriter, realm string) {
	w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
	w.WriteHeader(http.StatusUnauthorized)
}
