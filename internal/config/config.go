package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration for beehive-proxy.
type Config struct {
	TargetURL          string
	ListenAddr         string
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	MaxRequestsPerIP   int
	RateLimitWindow    time.Duration
	CBThreshold        int           // circuit-breaker failure threshold
	CBCooldown         time.Duration // circuit-breaker cooldown period
}

// FromEnv builds a Config from environment variables, applying defaults
// where values are absent. Returns an error when required fields are
// missing or values are invalid.
func FromEnv() (*Config, error) {
	target := envString("TARGET_URL", "")
	if target == "" {
		return nil, errors.New("TARGET_URL is required")
	}

	readTimeout, err := envDuration("READ_TIMEOUT_MS", 30000)
	if err != nil {
		return nil, fmt.Errorf("READ_TIMEOUT_MS: %w", err)
	}
	if readTimeout <= 0 {
		return nil, errors.New("READ_TIMEOUT_MS must be positive")
	}

	writeTimeout, err := envDuration("WRITE_TIMEOUT_MS", 30000)
	if err != nil {
		return nil, fmt.Errorf("WRITE_TIMEOUT_MS: %w", err)
	}
	if writeTimeout <= 0 {
		return nil, errors.New("WRITE_TIMEOUT_MS must be positive")
	}

	maxReq, err := envInt("MAX_REQUESTS_PER_IP", 100)
	if err != nil {
		return nil, fmt.Errorf("MAX_REQUESTS_PER_IP: %w", err)
	}

	rateLimitWindow, err := envDuration("RATE_LIMIT_WINDOW_MS", 1000)
	if err != nil {
		return nil, fmt.Errorf("RATE_LIMIT_WINDOW_MS: %w", err)
	}

	cbThreshold, err := envInt("CB_THRESHOLD", 5)
	if err != nil {
		return nil, fmt.Errorf("CB_THRESHOLD: %w", err)
	}

	cbCooldown, err := envDuration("CB_COOLDOWN_MS", 10000)
	if err != nil {
		return nil, fmt.Errorf("CB_COOLDOWN_MS: %w", err)
	}

	return &Config{
		TargetURL:        target,
		ListenAddr:       envString("LISTEN_ADDR", ":8080"),
		ReadTimeout:      readTimeout,
		WriteTimeout:     writeTimeout,
		MaxRequestsPerIP: maxReq,
		RateLimitWindow:  rateLimitWindow,
		CBThreshold:     cbThreshold,
		CBCooldown:      cbCooldown,
	}, nil
}

func envString(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid integer %q", v)
	}
	return n, nil
}

func envDuration(key string, defaultMS int) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return time.Duration(defaultMS) * time.Millisecond, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid integer %q", v)
	}
	return time.Duration(n) * time.Millisecond, nil
}
