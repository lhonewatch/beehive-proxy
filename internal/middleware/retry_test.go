package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

type stubTransport struct {
	callCount atomic.Int32
	responses []*http.Response
	errs      []error
}

func (s *stubTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	idx := int(s.callCount.Add(1)) - 1
	if idx >= len(s.responses) {
		idx = len(s.responses) - 1
	}
	return s.responses[idx], s.errs[idx]
}

func newStubReq(t *testing.T) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	return req
}

func TestRetryTransport_SucceedsFirstAttempt(t *testing.T) {
	stub := &stubTransport{
		responses: []*http.Response{{StatusCode: http.StatusOK}},
		errs:      []error{nil},
	}
	rt := NewRetryTransport(stub, RetryConfig{MaxAttempts: 3, Delay: time.Millisecond})
	resp, err := rt.RoundTrip(newStubReq(t))
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %v %v", resp, err)
	}
	if stub.callCount.Load() != 1 {
		t.Errorf("expected 1 call, got %d", stub.callCount.Load())
	}
}

func TestRetryTransport_RetriesOn5xx(t *testing.T) {
	stub := &stubTransport{
		responses: []*http.Response{
			{StatusCode: http.StatusServiceUnavailable},
			{StatusCode: http.StatusServiceUnavailable},
			{StatusCode: http.StatusOK},
		},
		errs: []error{nil, nil, nil},
	}
	rt := NewRetryTransport(stub, RetryConfig{MaxAttempts: 3, Delay: time.Millisecond})
	resp, err := rt.RoundTrip(newStubReq(t))
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected eventual 200 OK, got %v %v", resp, err)
	}
	if stub.callCount.Load() != 3 {
		t.Errorf("expected 3 calls, got %d", stub.callCount.Load())
	}
}

func TestRetryTransport_RetriesOnError(t *testing.T) {
	transportErr := errors.New("connection refused")
	stub := &stubTransport{
		responses: []*http.Response{nil, nil, {StatusCode: http.StatusOK}},
		errs:      []error{transportErr, transportErr, nil},
	}
	rt := NewRetryTransport(stub, RetryConfig{MaxAttempts: 3, Delay: time.Millisecond})
	resp, err := rt.RoundTrip(newStubReq(t))
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected recovery after retries, got %v %v", resp, err)
	}
}

func TestRetryTransport_ExhaustsAttempts(t *testing.T) {
	stub := &stubTransport{
		responses: []*http.Response{
			{StatusCode: http.StatusBadGateway},
			{StatusCode: http.StatusBadGateway},
			{StatusCode: http.StatusBadGateway},
		},
		errs: []error{nil, nil, nil},
	}
	rt := NewRetryTransport(stub, RetryConfig{MaxAttempts: 3, Delay: time.Millisecond})
	resp, _ := rt.RoundTrip(newStubReq(t))
	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("expected 502 after exhausted retries")
	}
	if stub.callCount.Load() != 3 {
		t.Errorf("expected 3 calls, got %d", stub.callCount.Load())
	}
}

func TestRetryTransport_ExhaustsAttemptsOnError(t *testing.T) {
	transportErr := errors.New("dial timeout")
	stub := &stubTransport{
		responses: []*http.Response{nil, nil, nil},
		errs:      []error{transportErr, transportErr, transportErr},
	}
	rt := NewRetryTransport(stub, RetryConfig{MaxAttempts: 3, Delay: time.Millisecond})
	_, err := rt.RoundTrip(newStubReq(t))
	if err == nil {
		t.Fatal("expected error after exhausted retries, got nil")
	}
	if stub.callCount.Load() != 3 {
		t.Errorf("expected 3 calls, got %d", stub.callCount.Load())
	}
}
