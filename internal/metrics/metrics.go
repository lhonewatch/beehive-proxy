package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestDuration tracks latency of proxied requests as a histogram.
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "beehive",
			Subsystem: "proxy",
			Name:      "request_duration_seconds",
			Help:      "Histogram of proxied request latencies in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "target_host", "status_code"},
	)

	// RequestsTotal counts the total number of proxied requests.
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "beehive",
			Subsystem: "proxy",
			Name:      "requests_total",
			Help:      "Total number of proxied requests.",
		},
		[]string{"method", "target_host", "status_code"},
	)

	// ActiveRequests tracks currently in-flight proxied requests.
	ActiveRequests = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "beehive",
			Subsystem: "proxy",
			Name:      "active_requests",
			Help:      "Number of proxied requests currently in flight.",
		},
	)
)

// ObserveRequest records duration, increments the total counter, and
// decrements the active gauge for a completed proxied request.
func ObserveRequest(method, targetHost, statusCode string, durationSeconds float64) {
	labels := prometheus.Labels{
		"method":      method,
		"target_host": targetHost,
		"status_code": statusCode,
	}
	RequestDuration.With(labels).Observe(durationSeconds)
	RequestsTotal.With(labels).Inc()
	ActiveRequests.Dec()
}
