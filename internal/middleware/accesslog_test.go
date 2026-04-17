package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/patrickward/beehive-proxy/internal/middleware"
)

func buildAccessLogger(t *testing.T) (*zap.Logger, *observer.ObservedLogs) {
	t.Helper()
	core, logs := observer.New(zapcore.InfoLevel)
	return zap.New(core), logs
}

func TestAccessLog_LogsMethod(t *testing.T) {
	logger, logs := buildAccessLogger(t)
	h := middleware.NewAccessLog(middleware.AccessLogOptions{Logger: logger})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }),
	)
	req := httptest.NewRequest(http.MethodPost, "/api", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if logs.Len() != 1 {
		t.Fatalf("expected 1 log entry, got %d", logs.Len())
	}
	entry := logs.All()[0]
	if entry.ContextMap()["method"] != "POST" {
		t.Errorf("expected method POST, got %v", entry.ContextMap()["method"])
	}
}

func TestAccessLog_LogsStatusCode(t *testing.T) {
	logger, logs := buildAccessLogger(t)
	h := middleware.NewAccessLog(middleware.AccessLogOptions{Logger: logger})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(404) }),
	)
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	fields := logs.All()[0].ContextMap()
	if fields["status"] != int64(404) {
		t.Errorf("expected status 404, got %v", fields["status"])
	}
}

func TestAccessLog_SkipsConfiguredPaths(t *testing.T) {
	logger, logs := buildAccessLogger(t)
	h := middleware.NewAccessLog(middleware.AccessLogOptions{
		Logger:    logger,
		SkipPaths: []string{"/healthz"},
	})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }),
	)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if logs.Len() != 0 {
		t.Errorf("expected no log entries for skipped path, got %d", logs.Len())
	}
}

func TestAccessLog_LogsRequestID(t *testing.T) {
	logger, logs := buildAccessLogger(t)
	h := middleware.NewAccessLog(middleware.AccessLogOptions{Logger: logger})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }),
	)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Id", "abc-123")
	h.ServeHTTP(httptest.NewRecorder(), req)

	fields := logs.All()[0].ContextMap()
	if fields["request_id"] != "abc-123" {
		t.Errorf("expected request_id abc-123, got %v", fields["request_id"])
	}
}
