package middleware

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

// RequestSizeHistogram records a Prometheus histogram of incoming request body
// sizes (in bytes). Requests without a Content-Length header are counted in the
// zero bucket.
type requestSizeHistogram struct {
	histogram prometheus.Histogram
}

// NewRequestSizeHistogram returns middleware that observes request body sizes
// using the provided Prometheus histogram. If hist is nil a default histogram
// registered with prometheus.DefaultRegisterer is used.
func NewRequestSizeHistogram(hist prometheus.Histogram) func(http.Handler) http.Handler {
	if hist == nil {
		hist = prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "beehive",
			Name:      "request_size_bytes",
			Help:      "Distribution of incoming HTTP request body sizes in bytes.",
			Buckets:   prometheus.ExponentialBuckets(64, 4, 8), // 64 B … ~1 MB
		})
		_ = prometheus.Register(hist)
	}

	m := &requestSizeHistogram{histogram: hist}
	return m.handler
}

func (m *requestSizeHistogram) handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		size := float64(0)
		if cl := r.Header.Get("Content-Length"); cl != "" {
			if n, err := strconv.ParseFloat(cl, 64); err == nil && n > 0 {
				size = n
			}
		}
		m.histogram.Observe(size)
		next.ServeHTTP(w, r)
	})
}
