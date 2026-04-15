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
	TargetURL       string
	ListenAddr      string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	RateLimit       int           // requests per window per IP (0 = disabled)
	RateLimitWindow time.Duration // window duration for rate limiting
}

// FromEnv loads configuration from environment variables.
func FromEnv() (*Config, error) {
	target := envString("TARGET_URL", "")
	if target == "" {
		return nil, errors.New("TARGET_URL is required")
	}

	readSec := envInt("READ_TIMEOUT_SECONDS", 30)
	if readSec <= 0 {
		return nil, fmt.Errorf("READ_TIMEOUT_SECONDS must be positive, got %d", readSec)
	}

	writeSec := envInt("WRITE_TIMEOUT_SECONDS", 30)
	if writeSec <= 0 {
		return nil, fmt.Errorf("WRITE_TIMEOUT_SECONDS must be positive, got %d", writeSec)
	}

	rateLimitWindowSec := envInt("RATE_LIMIT_WINDOW_SECONDS", 60)
	if rateLimitWindowSec <= 0 {
		return nil, fmt.Errorf("RATE_LIMIT_WINDOW_SECONDS must be positive, got %d", rateLimitWindowSec)
	}

	return &Config{
		TargetURL:       target,
		ListenAddr:      envString("LISTEN_ADDR", ":8080"),
		ReadTimeout:     time.Duration(readSec) * time.Second,
		WriteTimeout:    time.Duration(writeSec) * time.Second,
		RateLimit:       envInt("RATE_LIMIT", 0),
		RateLimitWindow: time.Duration(rateLimitWindowSec) * time.Second,
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
