package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/beehive-proxy/internal/middleware"
)

func idempotencyBackend(calls *atomic.Int32) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("X-Call", "yes")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created")) //nolint:errcheck
	})
}

func buildIdempotency(ttl time.Duration) (http.Handler, *atomic.Int32) {
	var calls atomic.Int32
	h := middleware.NewRequestIdempotency("Idempotency-Key", ttl)(idempotencyBackend(&calls))
	return h, &calls
}

func TestRequestIdempotency_PassesThroughWithoutKey(t *testing.T) {
	h, calls := buildIdempotency(time.Minute)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if calls.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", calls.Load())
	}
}

func TestRequestIdempotency_CachesResponseForSameKey(t *testing.T) {
	h, calls := buildIdempotency(time.Minute)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Idempotency-Key", "key-abc")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("iteration %d: expected 201, got %d", i, rec.Code)
		}
		body, _ := io.ReadAll(rec.Body)
		if string(body) != "created" {
			t.Fatalf("iteration %d: unexpected body %q", i, body)
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("expected backend called once, got %d", calls.Load())
	}
}

func TestRequestIdempotency_SetsReplayedHeader(t *testing.T) {
	h, _ := buildIdempotency(time.Minute)
	send := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Idempotency-Key", "key-replay")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}
	send() // prime cache
	rec := send()
	if rec.Header().Get("X-Idempotency-Replayed") != "true" {
		t.Fatal("expected X-Idempotency-Replayed: true on replay")
	}
}

func TestRequestIdempotency_ExpiresTTL(t *testing.T) {
	h, calls := buildIdempotency(10 * time.Millisecond)
	send := func() {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Idempotency-Key", "key-ttl")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
	}
	send()
	time.Sleep(30 * time.Millisecond)
	send()
	if calls.Load() != 2 {
		t.Fatalf("expected 2 backend calls after TTL expiry, got %d", calls.Load())
	}
}

func TestRequestIdempotency_IndependentKeys(t *testing.T) {
	h, calls := buildIdempotency(time.Minute)
	for _, key := range []string{"k1", "k2", "k3"} {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Idempotency-Key", key)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
	}
	if calls.Load() != 3 {
		t.Fatalf("expected 3 backend calls for 3 distinct keys, got %d", calls.Load())
	}
}
