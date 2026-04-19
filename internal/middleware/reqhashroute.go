package middleware

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"
)

// HashRouteOptions configures consistent hash-based routing header injection.
type HashRouteOptions struct {
	// Header to read the hash key from (e.g. "X-User-ID").
	KeyHeader string
	// OutputHeader is the header written with the computed bucket index.
	OutputHeader string
	// Buckets is the number of hash buckets (shards).
	Buckets uint32
	// Fallback is the value written when KeyHeader is absent.
	Fallback string
}

// NewHashRoute returns middleware that computes a consistent hash bucket from
// a request header and injects the result as a downstream header. Useful for
// sticky routing decisions made by an upstream load-balancer or proxy.
func NewHashRoute(opts HashRouteOptions) func(http.Handler) http.Handler {
	if opts.OutputHeader == "" {
		opts.OutputHeader = "X-Route-Bucket"
	}
	if opts.Buckets == 0 {
		opts.Buckets = 10
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := strings.TrimSpace(r.Header.Get(opts.KeyHeader))
			if key == "" {
				if opts.Fallback != "" {
					r.Header.Set(opts.OutputHeader, opts.Fallback)
				}
				next.ServeHTTP(w, r)
				return
			}
			bucket := hashBucket(key, opts.Buckets)
			r.Header.Set(opts.OutputHeader, fmt.Sprintf("%d", bucket))
			next.ServeHTTP(w, r)
		})
	}
}

func hashBucket(key string, buckets uint32) uint32 {
	h := md5.Sum([]byte(key))
	// Use first 4 bytes as uint32.
	v := uint32(h[0])<<24 | uint32(h[1])<<16 | uint32(h[2])<<8 | uint32(h[3])
	return v % buckets
}
