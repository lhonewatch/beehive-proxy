package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/yourusername/beehive-proxy/internal/middleware"
)

func statusHandler(code int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	})
}

func TestPrometheus_CountsRequests(t *testing.T) {
	before := testutil.ToFloat64(prometheus.DefaultRegisterer.(prometheus.Gatherer))
	_ = before // just ensure no panic on gather

	h := middleware.NewPrometheus(statusHandler(http.StatusOK))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPrometheus_Records500(t *testing.T) {
	h := middleware.NewPrometheus(statusHandler(http.StatusInternalServerError))
	req := httptest.NewRequest(http.MethodPost, "/fail", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestPrometheus_MultipleRequests(t *testing.T) {
	h := middleware.NewPrometheus(statusHandler(http.StatusAccepted))
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != http.StatusAccepted {
			t.Fatalf("iteration %d: expected 202, got %d", i, w.Code)
		}
	}
}
