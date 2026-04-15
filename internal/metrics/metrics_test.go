package metrics_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/your-org/beehive-proxy/internal/metrics"
)

func TestObserveRequest_IncrementsCounter(t *testing.T) {
	// Reset state by reading current value before the call.
	before := testutil.ToFloat64(metrics.RequestsTotal.With(prometheus.Labels{
		"method":      "GET",
		"target_host": "example.com",
		"status_code": "200",
	}))

	metrics.ActiveRequests.Inc() // simulate an in-flight request
	metrics.ObserveRequest("GET", "example.com", "200", 0.042)

	after := testutil.ToFloat64(metrics.RequestsTotal.With(prometheus.Labels{
		"method":      "GET",
		"target_host": "example.com",
		"status_code": "200",
	}))

	if after-before != 1 {
		t.Errorf("expected counter to increase by 1, got delta %v", after-before)
	}
}

func TestObserveRequest_DecrementsActiveGauge(t *testing.T) {
	metrics.ActiveRequests.Inc()
	metrics.ActiveRequests.Inc()

	before := testutil.ToFloat64(metrics.ActiveRequests)
	metrics.ObserveRequest("POST", "api.internal", "500", 1.5)
	after := testutil.ToFloat64(metrics.ActiveRequests)

	if before-after != 1 {
		t.Errorf("expected active gauge to decrease by 1, got delta %v", before-after)
	}
}

func TestObserveRequest_RecordsDuration(t *testing.T) {
	metrics.ActiveRequests.Inc()
	// Ensure no panic and histogram accepts the observation.
	metrics.ObserveRequest("DELETE", "storage.svc", "204", 0.001)
}
