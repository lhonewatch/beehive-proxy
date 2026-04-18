package middleware

import (
	"net/http"
	"strings"
)

// GeoFenceOptions configures the GeoFence middleware.
type GeoFenceOptions struct {
	// Mode is either "allowlist" or "blocklist".
	Mode string
	// Countries is a list of ISO 3166-1 alpha-2 country codes.
	Countries []string
	// CountryHeader is the request header that carries the country code.
	// Defaults to "X-Country-Code".
	CountryHeader string
}

// NewGeoFence returns a middleware that allows or blocks requests based on a
// country code header set by an upstream CDN or load balancer.
func NewGeoFence(opts GeoFenceOptions) func(http.Handler) http.Handler {
	if opts.CountryHeader == "" {
		opts.CountryHeader = "X-Country-Code"
	}

	set := make(map[string]struct{}, len(opts.Countries))
	for _, c := range opts.Countries {
		set[strings.ToUpper(c)] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			country := strings.ToUpper(strings.TrimSpace(r.Header.Get(opts.CountryHeader)))

			_, found := set[country]

			switch strings.ToLower(opts.Mode) {
			case "allowlist":
				if !found {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
			case "blocklist":
				if found {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
