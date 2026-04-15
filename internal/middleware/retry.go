package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// RetryConfig holds configuration for the retry middleware.
type RetryConfig struct {
	MaxAttempts int
	Delay       time.Duration
	Logger      *slog.Logger
}

// retryTransport wraps an http.RoundTripper with retry logic.
type retryTransport struct {
	base        http.RoundTripper
	maxAttempts int
	delay       time.Duration
	logger      *slog.Logger
}

// NewRetryTransport returns an http.RoundTripper that retries failed requests.
func NewRetryTransport(base http.RoundTripper, cfg RetryConfig) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 3
	}
	if cfg.Delay <= 0 {
		cfg.Delay = 100 * time.Millisecond
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return &retryTransport{
		base:        base,
		maxAttempts: cfg.MaxAttempts,
		delay:       cfg.Delay,
		logger:      cfg.Logger,
	}
}

// RoundTrip executes the request, retrying on 5xx or transport errors.
func (rt *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)
	for attempt := 1; attempt <= rt.maxAttempts; attempt++ {
		resp, err = rt.base.RoundTrip(req)
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}
		if attempt < rt.maxAttempts {
			rt.logger.Warn("retrying request",
				"attempt", attempt,
				"max", rt.maxAttempts,
				"url", req.URL.String(),
				"error", err,
			)
			time.Sleep(rt.delay)
		}
	}
	return resp, err
}
