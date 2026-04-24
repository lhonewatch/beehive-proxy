package middleware

import (
	"net/http"
	"runtime"
	"strconv"
)

// RequestMetadata injects server-side metadata headers into each proxied request.
// It adds the Go runtime version, number of logical CPUs, and a configurable
// service name so downstream services can identify the originating proxy.
type requestMetadataOptions struct {
	ServiceName    string
	ServiceVersion string
	RequestHeader  string
}

// RequestMetadataOption configures NewRequestMetadata.
type RequestMetadataOption func(*requestMetadataOptions)

// WithMetadataServiceName sets the X-Service-Name header value.
func WithMetadataServiceName(name string) RequestMetadataOption {
	return func(o *requestMetadataOptions) { o.ServiceName = name }
}

// WithMetadataServiceVersion sets the X-Service-Version header value.
func WithMetadataServiceVersion(version string) RequestMetadataOption {
	return func(o *requestMetadataOptions) { o.ServiceVersion = version }
}

// WithMetadataRequestHeader overrides the header name used to carry the
// service name (default: X-Service-Name).
func WithMetadataRequestHeader(header string) RequestMetadataOption {
	return func(o *requestMetadataOptions) { o.RequestHeader = header }
}

// NewRequestMetadata returns a middleware that stamps outbound requests with
// runtime and service identity headers.
func NewRequestMetadata(next http.Handler, opts ...RequestMetadataOption) http.Handler {
	o := &requestMetadataOptions{
		ServiceName:    "beehive-proxy",
		ServiceVersion: "dev",
		RequestHeader:  "X-Service-Name",
	}
	for _, opt := range opts {
		opt(o)
	}

	goVersion := runtime.Version()
	numCPU := strconv.Itoa(runtime.NumCPU())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.Clone(r.Context())
		r.Header.Set(o.RequestHeader, o.ServiceName)
		r.Header.Set("X-Service-Version", o.ServiceVersion)
		r.Header.Set("X-Runtime-Version", goVersion)
		r.Header.Set("X-Runtime-CPUs", numCPU)
		next.ServeHTTP(w, r)
	})
}
