package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/beehive-proxy/internal/middleware"
)

func okSamplerHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestRequestSampler_RateZeroDropsAll(t *testing.T) {
	h := middleware.NewRequestSampler(0)(http.HandlerFunc(okSamplerHandler))
	for i := 0; i < 10; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}
	}
}

func TestRequestSampler_RateOneAllowsAll(t *testing.T) {
	h := middleware.NewRequestSampler(1.0)(http.HandlerFunc(okSamplerHandler))
	for i := 0; i < 10; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	}
}

func TestRequestSampler_PartialRatePassesSome(t *testing.T) {
	h := middleware.NewRequestSampler(0.5)(http.HandlerFunc(okSamplerHandler))
	ok, dropped := 0, 0
	for i := 0; i < 1000; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rec.Code == http.StatusOK {
			ok++
		} else {
			dropped++
		}
	}
	if ok == 0 || dropped == 0 {
		t.Errorf("expected mix of passed/dropped, got ok=%d dropped=%d", ok, dropped)
	}
}

func TestRequestSampler_DroppedReturns204(t *testing.T) {
	// rate=0 guarantees drop
	h := middleware.NewRequestSampler(0)(http.HandlerFunc(okSamplerHandler))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}
