package tracing_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beehive-proxy/internal/tracing"
)

func TestDefaultTracerFunc_GeneratesID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	id := tracing.DefaultTracerFunc(req)
	if id == "" {
		t.Fatal("expected a non-empty trace ID to be generated")
	}
}

func TestDefaultTracerFunc_ReusesExistingID(t *testing.T) {
	const existing = "abc123"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(tracing.TraceIDHeader, existing)

	id := tracing.DefaultTracerFunc(req)
	if id != existing {
		t.Fatalf("expected %q, got %q", existing, id)
	}
}

func TestDefaultTracerFunc_UniqueIDs(t *testing.T) {
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)

	id1 := tracing.DefaultTracerFunc(req1)
	id2 := tracing.DefaultTracerFunc(req2)

	if id1 == id2 {
		t.Fatalf("expected unique trace IDs, both were %q", id1)
	}
}

func TestDefaultTracerFunc_EmptyHeaderGeneratesID(t *testing.T) {
	// Ensure that an explicitly empty header value is treated as absent
	// and a new trace ID is generated rather than propagating an empty string.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(tracing.TraceIDHeader, "")

	id := tracing.DefaultTracerFunc(req)
	if id == "" {
		t.Fatal("expected a non-empty trace ID when header is explicitly empty")
	}
}
