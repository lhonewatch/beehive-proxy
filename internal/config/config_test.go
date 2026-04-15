package config

import (
	"testing"
	"time"
)

func setEnv(t *testing.T, key, value string) {
	t.Helper()
	t.Setenv(key, value)
}

func TestFromEnv_MissingTargetURL(t *testing.T) {
	t.Setenv("BEEHIVE_TARGET_URL", "")
	_, err := FromEnv()
	if err == nil {
		t.Fatal("expected error for missing BEEHIVE_TARGET_URL")
	}
}

func TestFromEnv_Defaults(t *testing.T) {
	setEnv(t, "BEEHIVE_TARGET_URL", "http://localhost:9090")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ListenAddr != ":8080" {
		t.Errorf("expected :8080, got %s", cfg.ListenAddr)
	}
	if cfg.ReadTimeout != 30*time.Second {
		t.Errorf("expected 30s read timeout, got %v", cfg.ReadTimeout)
	}
	if cfg.RetryMaxAttempts != 3 {
		t.Errorf("expected 3 retry attempts, got %d", cfg.RetryMaxAttempts)
	}
	if cfg.RetryDelay != 100*time.Millisecond {
		t.Errorf("expected 100ms retry delay, got %v", cfg.RetryDelay)
	}
	if cfg.MetricsPath != "/metrics" {
		t.Errorf("expected /metrics, got %s", cfg.MetricsPath)
	}
}

func TestFromEnv_CustomValues(t *testing.T) {
	setEnv(t, "BEEHIVE_TARGET_URL", "http://backend:8081")
	setEnv(t, "BEEHIVE_LISTEN_ADDR", ":9000")
	setEnv(t, "BEEHIVE_RATE_LIMIT", "50")
	setEnv(t, "BEEHIVE_RETRY_MAX_ATTEMPTS", "5")
	setEnv(t, "BEEHIVE_RETRY_DELAY", "200ms")

	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ListenAddr != ":9000" {
		t.Errorf("expected :9000, got %s", cfg.ListenAddr)
	}
	if cfg.RateLimit != 50 {
		t.Errorf("expected rate limit 50, got %d", cfg.RateLimit)
	}
	if cfg.RetryMaxAttempts != 5 {
		t.Errorf("expected 5 retry attempts, got %d", cfg.RetryMaxAttempts)
	}
	if cfg.RetryDelay != 200*time.Millisecond {
		t.Errorf("expected 200ms retry delay, got %v", cfg.RetryDelay)
	}
}

func TestFromEnv_InvalidTimeout(t *testing.T) {
	setEnv(t, "BEEHIVE_TARGET_URL", "http://localhost:9090")
	setEnv(t, "BEEHIVE_READ_TIMEOUT", "not-a-duration")
	_, err := FromEnv()
	if err == nil {
		t.Fatal("expected error for invalid BEEHIVE_READ_TIMEOUT")
	}
}

func TestFromEnv_InvalidRetryDelay(t *testing.T) {
	setEnv(t, "BEEHIVE_TARGET_URL", "http://localhost:9090")
	setEnv(t, "BEEHIVE_RETRY_DELAY", "bad")
	_, err := FromEnv()
	if err == nil {
		t.Fatal("expected error for invalid BEEHIVE_RETRY_DELAY")
	}
}

func TestFromEnv_InvalidTargetURL(t *testing.T) {
	setEnv(t, "BEEHIVE_TARGET_URL", "://bad-url")
	_, err := FromEnv()
	if err == nil {
		t.Fatal("expected error for invalid BEEHIVE_TARGET_URL")
	}
}
