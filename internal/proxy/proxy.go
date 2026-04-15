package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/beehive-proxy/internal/metrics"
	"github.com/beehive-proxy/internal/tracing"
)

// Handler is a reverse proxy handler that records metrics and traces.
type Handler struct {
	target   *url.URL
	proxy    *httputil.ReverseProxy
	tracerFn tracing.TracerFunc
}

// NewHandler creates a new reverse proxy Handler targeting the given URL.
func NewHandler(target *url.URL, tracerFn tracing.TracerFunc) *Handler {
	rp := httputil.NewSingleHostReverseProxy(target)
	return &Handler{
		target:   target,
		proxy:    rp,
		tracerFn: tracerFn,
	}
}

// ServeHTTP proxies the request to the upstream target, recording
// latency and request counts via Prometheus metrics.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	traceID := h.tracerFn(r)
	r.Header.Set(tracing.TraceIDHeader, traceID)

	rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
	h.proxy.ServeHTTP(rw, r)

	duration := time.Since(start).Seconds()
	metrics.ObserveRequest(r.Method, rw.statusCode, duration)
}
