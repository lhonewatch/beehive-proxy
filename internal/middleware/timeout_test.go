package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/beehive-proxy/internal/middleware"
)

func fastHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func slowHandler(delay time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(delay):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("late"))
		case <-r.Context().Done():
			// context cancelled — do nothing
		}
	}
}

func TestTimeout_AllowsFastHandler(t *testing.T) {
	h := middleware.NewTimeout(100 * time.Millisecond)(http.HandlerFunc(fastHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "ok" {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestTimeout_BlocksSlowHandler(t *testing.T) {
	h := middleware.NewTimeout(50 * time.Millisecond)(slowHandler(200 * time.Millisecond))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d", rec.Code)
	}
}

func TestTimeout_PropagatesCancellation(t *testing.T) {
	cancelled := make(chan struct{})
	h := middleware.NewTimeout(30*time.Millisecond)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		close(cancelled)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	select {
	case <-cancelled:
		// context was cancelled as expected
	case <-time.After(200 * time.Millisecond):
		t.Fatal("context was not cancelled within expected window")
	}
}
