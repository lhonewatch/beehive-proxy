package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/sn3d/beehive-proxy/internal/middleware"
)

func captureTraceHandler(t *testing.T, traceID, spanID, traceStart *string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*traceID = r.Header.Get(middleware.TraceIDHeader)
		*spanID = r.Header.Get(middleware.SpanIDHeader)
		*traceStart = r.Header.Get(middleware.TraceStartHeader)
		w.WriteHeader(http.StatusOK)
	})
}

func TestRequestTrace_SetsTraceAndSpanIDWhenAbsent(t *testing.T) {
	var traceID, spanID, traceStart string
	h := middleware.NewRequestTrace(captureTraceHandler(t, &traceID, &spanID, &traceStart))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if traceID == "" {
		t.Fatal("expected X-Trace-ID to be set")
	}
	if spanID == "" {
		t.Fatal("expected X-Span-ID to be set")
	}
	if traceStart == "" {
		t.Fatal("expected X-Trace-Start to be set")
	}
}

func TestRequestTrace_PreservesExistingTraceID(t *testing.T) {
	const existing = "my-trace-id"
	var traceID, spanID, traceStart string
	h := middleware.NewRequestTrace(captureTraceHandler(t, &traceID, &spanID, &traceStart))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(middleware.TraceIDHeader, existing)
	h.ServeHTTP(rec, req)

	if traceID != existing {
		t.Fatalf("expected trace ID %q, got %q", existing, traceID)
	}
	// X-Trace-Start must NOT be set when trace ID was already present.
	if traceStart != "" {
		t.Fatalf("expected X-Trace-Start to be empty, got %q", traceStart)
	}
}

func TestRequestTrace_GeneratesUniqueSpanIDPerRequest(t *testing.T) {
	spans := make([]string, 3)
	for i := range spans {
		var traceID, spanID, traceStart string
		h := middleware.NewRequestTrace(captureTraceHandler(t, &traceID, &spanID, &traceStart))
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		h.ServeHTTP(rec, req)
		spans[i] = spanID
	}
	if spans[0] == spans[1] || spans[1] == spans[2] {
		t.Fatal("expected unique span IDs across requests")
	}
}

func TestRequestTrace_PropagatesHeadersToResponse(t *testing.T) {
	h := middleware.NewRequestTrace(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Header().Get(middleware.TraceIDHeader) == "" {
		t.Error("expected X-Trace-ID in response headers")
	}
	if rec.Header().Get(middleware.SpanIDHeader) == "" {
		t.Error("expected X-Span-ID in response headers")
	}
}

func TestRequestTrace_TraceStartIsRecentTimestamp(t *testing.T) {
	var traceID, spanID, traceStart string
	before := time.Now()
	h := middleware.NewRequestTrace(captureTraceHandler(t, &traceID, &spanID, &traceStart))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)
	after := time.Now()

	nano, err := strconv.ParseInt(traceStart, 10, 64)
	if err != nil {
		t.Fatalf("X-Trace-Start is not a valid int64: %v", err)
	}
	ts := time.Unix(0, nano)
	if ts.Before(before) || ts.After(after) {
		t.Errorf("X-Trace-Start %v not within expected range [%v, %v]", ts, before, after)
	}
}
