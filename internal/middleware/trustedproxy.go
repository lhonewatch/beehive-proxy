package middleware

import (
	"net"
	"net/http"
	"strings"
)

// TrustedProxyOptions configures the trusted proxy middleware.
type TrustedProxyOptions struct {
	// TrustedCIDRs is the list of CIDR ranges considered trusted proxies.
	TrustedCIDRs []string
}

type trustedProxy struct {
	nets []*net.IPNet
}

// NewTrustedProxy returns middleware that rewrites RemoteAddr using
// X-Forwarded-For only when the immediate peer is within a trusted CIDR.
func NewTrustedProxy(opts TrustedProxyOptions) func(http.Handler) http.Handler {
	tp := &trustedProxy{}
	for _, cidr := range opts.TrustedCIDRs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil {
			tp.nets = append(tp.nets, ipNet)
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if tp.isTrusted(peerIP(r.RemoteAddr)) {
				if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
					parts := strings.Split(xff, ",")
					client := strings.TrimSpace(parts[0])
					if net.ParseIP(client) != nil {
						r = r.Clone(r.Context())
						r.RemoteAddr = client
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (tp *trustedProxy) isTrusted(ip net.IP) bool {
	if ip == nil {
		return false
	}
	for _, n := range tp.nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

func peerIP(remoteAddr string) net.IP {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	return net.ParseIP(strings.TrimSpace(host))
}
