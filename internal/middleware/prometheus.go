package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "beehive",
		Name:      "request_duration_seconds",
		Help:      "Histogram of proxied request latencies.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "status"})

	requestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "beehive",
		Name:      "requests_total",
		Help:      "Total number of proxied requests.",
	}, []string{"method", "status"})

	activeRequests = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "beehive",
		Name:      "active_requests",
		Help:      "Number of requests currently being proxied.",
	})
)

// NewPrometheus returns middleware that records per-request Prometheus metrics.
func NewPrometheus(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		activeRequests.Inc()
		start := time.Now()

		rec := NewResponseRecorder(w)
		next.ServeHTTP(rec, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(rec.Status())

		activeRequests.Dec()
		requestDuration.WithLabelValues(r.Method, status).Observe(duration)
		requestsTotal.WithLabelValues(r.Method, status).Inc()
	})
}
