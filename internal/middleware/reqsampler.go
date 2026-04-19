package middleware

import (
	"math/rand"
	"net/http"
)

// NewRequestSampler returns a middleware that only forwards a percentage of
// requests to the next handler. Requests that are not sampled receive 204.
// rate must be between 0.0 and 1.0.
func NewRequestSampler(rate float64) func(http.Handler) http.Handler {
	if rate <= 0 {
		return func(_ http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})
		}
	}
	if rate >= 1.0 {
		return func(next http.Handler) http.Handler { return next }
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if rand.Float64() < rate {
				next.ServeHTTP(w, r)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	}
}
