package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration for beehive-proxy.
type Config struct {
	TargetURL       string
	ListenAddr      string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	MetricsPath     string
	RateLimit       int
	RateLimitWindow time.Duration
	CBThreshold     int
	CBCooldown      time.Duration
	RetryMaxAttempts int
	RetryDelay      time.Duration
}

// FromEnv reads configuration from environment variables, returning an error
// if required values are missing or invalid.
func FromEnv() (*Config, error) {
	target := envString("BEEHIVE_TARGET_URL", "")
	if target == "" {
		return nil, errors.New("BEEHIVE_TARGET_URL is required")
	}
	if _, err := url.ParseRequestURI(target); err != nil {
		return nil, fmt.Errorf("BEEHIVE_TARGET_URL is invalid: %w", err)
	}

	readTimeout, err := envDuration("BEEHIVE_READ_TIMEOUT", 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("BEEHIVE_READ_TIMEOUT: %w", err)
	}
	writeTimeout, err := envDuration("BEEHIVE_WRITE_TIMEOUT", 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("BEEHIVE_WRITE_TIMEOUT: %w", err)
	}
	idleTimeout, err := envDuration("BEEHIVE_IDLE_TIMEOUT", 60*time.Second)
	if err != nil {
		return nil, fmt.Errorf("BEEHIVE_IDLE_TIMEOUT: %w", err)
	}
	retryDelay, err := envDuration("BEEHIVE_RETRY_DELAY", 100*time.Millisecond)
	if err != nil {
		return nil, fmt.Errorf("BEEHIVE_RETRY_DELAY: %w", err)
	}
	cbCooldown, err := envDuration("BEEHIVE_CB_COOLDOWN", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("BEEHIVE_CB_COOLDOWN: %w", err)
	}
	rateLimitWindow, err := envDuration("BEEHIVE_RATE_LIMIT_WINDOW", time.Second)
	if err != nil {
		return nil, fmt.Errorf("BEEHIVE_RATE_LIMIT_WINDOW: %w", err)
	}

	return &Config{
		TargetURL:        target,
		ListenAddr:       envString("BEEHIVE_LISTEN_ADDR", ":8080"),
		ReadTimeout:      readTimeout,
		WriteTimeout:     writeTimeout,
		IdleTimeout:      idleTimeout,
		MetricsPath:      envString("BEEHIVE_METRICS_PATH", "/metrics"),
		RateLimit:        envInt("BEEHIVE_RATE_LIMIT", 100),
		RateLimitWindow:  rateLimitWindow,
		CBThreshold:      envInt("BEEHIVE_CB_THRESHOLD", 5),
		CBCooldown:       cbCooldown,
		RetryMaxAttempts: envInt("BEEHIVE_RETRY_MAX_ATTEMPTS", 3),
		RetryDelay:       retryDelay,
	}, nil
}

func envString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", v, err)
	}
	return d, nil
}
