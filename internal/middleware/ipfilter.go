package middleware

import (
	"net"
	"net/http"
	"strings"
)

// IPFilterMode controls whether the list is an allowlist or blocklist.
type IPFilterMode int

const (
	Allowlist IPFilterMode = iota
	Blocklist
)

// IPFilter holds the configuration for IP-based access control.
type IPFilter struct {
	mode    IPFilterMode
	nets    []*net.IPNet
	next    http.Handler
}

// NewIPFilter creates an IP filter middleware.
// cidrs is a list of CIDR strings (e.g. "192.168.1.0/24", "10.0.0.1/32").
// mode selects Allowlist (only listed IPs pass) or Blocklist (listed IPs are denied).
func NewIPFilter(mode IPFilterMode, cidrs []string) func(http.Handler) http.Handler {
	parsed := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		// Support bare IPs by appending /32 or /128.
		if !strings.Contains(cidr, "/") {
			if strings.Contains(cidr, ":") {
				cidr += "/128"
			} else {
				cidr += "/32"
			}
		}
		_, network, err := net.ParseCIDR(cidr)
		if err == nil {
			parsed = append(parsed, network)
		}
	}

	return func(next http.Handler) http.Handler {
		return &IPFilter{mode: mode, nets: parsed, next: next}
	}
}

func (f *IPFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ip := realIP(r)
	// If we cannot determine the client IP, deny the request to fail safely.
	if ip == nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	matched := f.matches(ip)

	if f.mode == Allowlist && !matched {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if f.mode == Blocklist && matched {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	f.next.ServeHTTP(w, r)
}

func (f *IPFilter) matches(ip net.IP) bool {
	for _, network := range f.nets {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// realIP extracts the client IP from X-Forwarded-For, X-Real-IP, or RemoteAddr.
func realIP(r *http.Request) net.IP {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		if ip := net.ParseIP(strings.TrimSpace(parts[0])); ip != nil {
			return ip
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if ip := net.ParseIP(strings.TrimSpace(xri)); ip != nil {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return net.ParseIP(r.RemoteAddr)
	}
	return net.ParseIP(host)
}
