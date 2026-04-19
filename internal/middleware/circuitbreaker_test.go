package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func errorHandler(code int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	})
}

func TestCircuitBreaker_AllowsWhenClosed(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)
	if !cb.Allow() {
		t.Fatal("expected circuit to be closed initially")
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)
	handler := cb.Middleware(errorHandler(http.StatusInternalServerError))

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// Next request should be rejected with 503
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestCircuitBreaker_ResetsAfterCooldown(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond)
	handler := cb.Middleware(errorHandler(http.StatusInternalServerError))

	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	time.Sleep(60 * time.Millisecond)

	// After cooldown a probe is allowed (half-open)
	if !cb.Allow() {
		t.Fatal("expected circuit to allow probe after cooldown")
	}
}

func TestCircuitBreaker_ClosesOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond)
	handler := cb.Middleware(errorHandler(http.StatusOK))

	cb.RecordFailure()
	cb.RecordFailure() // opens the circuit
	time.Sleep(60 * time.Millisecond)

	// Successful probe should close the circuit
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if !cb.Allow() {
		t.Fatal("expected circuit to be closed after successful probe")
	}
}

func TestCircuitBreaker_IndependentInstances(t *testing.T) {
	cb1 := NewCircuitBreaker(1, time.Minute)
	cb2 := NewCircuitBreaker(1, time.Minute)

	cb1.RecordFailure()

	if cb2.Allow() == false {
		t.Fatal("cb2 should be unaffected by cb1 failures")
	}
}

func TestCircuitBreaker_RemainsOpenDuringCooldown(t *testing.T) {
	cb := NewCircuitBreaker(2, 200*time.Millisecond)
	handler := cb.Middleware(errorHandler(http.StatusInternalServerError))

	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// Before cooldown expires, circuit should still be open
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 while circuit is open, got %d", rec.Code)
	}
}
