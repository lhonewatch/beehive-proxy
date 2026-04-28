package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"github.com/yourusername/beehive-proxy/internal/middleware"
)

func newSizeHistogram(t *testing.T) prometheus.Histogram {
	t.Helper()
	h := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "test_request_size_bytes",
		Help:    "test histogram",
		Buckets: prometheus.ExponentialBuckets(64, 4, 8),
	})
	return h
}

func collectHistogram(h prometheus.Histogram) *dto.Histogram {
	var m dto.Metric
	_ = h.Write(&m)
	return m.Histogram
}

var sizeOKHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestRequestSizeHistogram_ObservesZeroWhenNoContentLength(t *testing.T) {
	h := newSizeHistogram(t)
	handler := middleware.NewRequestSizeHistogram(h)(sizeOKHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	hist := collectHistogram(h)
	if got := hist.GetSampleCount(); got != 1 {
		t.Fatalf("expected 1 sample, got %d", got)
	}
	if got := hist.GetSampleSum(); got != 0 {
		t.Fatalf("expected sum 0, got %f", got)
	}
}

func TestRequestSizeHistogram_ObservesContentLength(t *testing.T) {
	h := newSizeHistogram(t)
	handler := middleware.NewRequestSizeHistogram(h)(sizeOKHandler)

	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	req.Header.Set("Content-Length", "512")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	hist := collectHistogram(h)
	if got := hist.GetSampleCount(); got != 1 {
		t.Fatalf("expected 1 sample, got %d", got)
	}
	if got := hist.GetSampleSum(); got != 512 {
		t.Fatalf("expected sum 512, got %f", got)
	}
}

func TestRequestSizeHistogram_AccumulatesMultipleRequests(t *testing.T) {
	h := newSizeHistogram(t)
	handler := middleware.NewRequestSizeHistogram(h)(sizeOKHandler)

	sizes := []string{"100", "200", "300"}
	for _, s := range sizes {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Content-Length", s)
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}

	hist := collectHistogram(h)
	if got := hist.GetSampleCount(); got != 3 {
		t.Fatalf("expected 3 samples, got %d", got)
	}
	if got := hist.GetSampleSum(); got != 600 {
		t.Fatalf("expected sum 600, got %f", got)
	}
}

func TestRequestSizeHistogram_IgnoresInvalidContentLength(t *testing.T) {
	h := newSizeHistogram(t)
	handler := middleware.NewRequestSizeHistogram(h)(sizeOKHandler)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Length", "not-a-number")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	hist := collectHistogram(h)
	if got := hist.GetSampleSum(); got != 0 {
		t.Fatalf("expected sum 0 for invalid header, got %f", got)
	}
}
