package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the runtime configuration for beehive-proxy.
type Config struct {
	// ListenAddr is the address the proxy listens on, e.g. ":8080".
	ListenAddr string

	// TargetURL is the backend URL to proxy requests to.
	TargetURL string

	// MetricsAddr is the address to expose Prometheus metrics on, e.g. ":9090".
	MetricsAddr string

	// ReadTimeout is the maximum duration for reading the entire request.
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out writes of the response.
	WriteTimeout time.Duration
}

// FromEnv builds a Config from environment variables with sensible defaults.
//
// Environment variables:
//   PROXY_LISTEN_ADDR  (default: ":8080")
//   PROXY_TARGET_URL   (required)
//   PROXY_METRICS_ADDR (default: ":9090")
//   PROXY_READ_TIMEOUT_S  (seconds, default: 30)
//   PROXY_WRITE_TIMEOUT_S (seconds, default: 30)
func FromEnv() (*Config, error) {
	target := os.Getenv("PROXY_TARGET_URL")
	if target == "" {
		return nil, fmt.Errorf("config: PROXY_TARGET_URL must be set")
	}

	readSecs, err := envInt("PROXY_READ_TIMEOUT_S", 30)
	if err != nil {
		return nil, err
	}

	writeSecs, err := envInt("PROXY_WRITE_TIMEOUT_S", 30)
	if err != nil {
		return nil, err
	}

	return &Config{
		ListenAddr:   envString("PROXY_LISTEN_ADDR", ":8080"),
		TargetURL:    target,
		MetricsAddr:  envString("PROXY_METRICS_ADDR", ":9090"),
		ReadTimeout:  time.Duration(readSecs) * time.Second,
		WriteTimeout: time.Duration(writeSecs) * time.Second,
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
		return 0, fmt.Errorf("config: %s must be an integer, got %q", key, v)
	}
	if n <= 0 {
		return 0, fmt.Errorf("config: %s must be positive, got %d", key, n)
	}
	return n, nil
}
