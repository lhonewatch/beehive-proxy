package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andriikushch/beehive-proxy/internal/middleware"
)

func buildSlowLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(buf, nil))
}

func slowOKHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestSlowLog_DoesNotLogFastRequest(t *testing.T) {
	var buf bytes.Buffer
	handler := middleware.NewSlowLog(500*time.Millisecond, buildSlowLogger(&buf))(http.HandlerFunc(slowOKHandler))
	req := httptest.NewRequest(http.MethodGet, "/fast", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if buf.Len() != 0 {
		t.Fatalf("expected no log output for fast request, got: %s", buf.String())
	}
}

func TestSlowLog_LogsSlowRequest(t *testing.T) {
	var buf bytes.Buffer
	threshold := 10 * time.Millisecond
	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	handler := middleware.NewSlowLog(threshold, buildSlowLogger(&buf))(slow)
	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if buf.Len() == 0 {
		t.Fatal("expected log output for slow request")
	}
	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON log: %v", err)
	}
	if entry["msg"] != "slow request" {
		t.Fatalf("unexpected msg: %v", entry["msg"])
	}
}

func TestSlowLog_IncludesTraceID(t *testing.T) {
	var buf bytes.Buffer
	threshold := 5 * time.Millisecond
	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(15 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	handler := middleware.NewSlowLog(threshold, buildSlowLogger(&buf))(slow)
	req := httptest.NewRequest(http.MethodGet, "/trace", nil)
	req.Header.Set("X-Trace-ID", "abc-123")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	var entry map[string]any
	_ = json.Unmarshal(buf.Bytes(), &entry)
	if entry["trace_id"] != "abc-123" {
		t.Fatalf("expected trace_id abc-123, got %v", entry["trace_id"])
	}
}
