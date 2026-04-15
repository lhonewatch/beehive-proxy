package config

import (
	"testing"
	"time"
)

func TestFromEnv_MissingTargetURL(t *testing.T) {
	t.Setenv("PROXY_TARGET_URL", "")

	_, err := FromEnv()
	if err == nil {
		t.Fatal("expected error when PROXY_TARGET_URL is not set, got nil")
	}
}

func TestFromEnv_Defaults(t *testing.T) {
	t.Setenv("PROXY_TARGET_URL", "http://backend:8081")
	t.Setenv("PROXY_LISTEN_ADDR", "")
	t.Setenv("PROXY_METRICS_ADDR", "")
	t.Setenv("PROXY_READ_TIMEOUT_S", "")
	t.Setenv("PROXY_WRITE_TIMEOUT_S", "")

	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ListenAddr != ":8080" {
		t.Errorf("ListenAddr: want :8080, got %s", cfg.ListenAddr)
	}
	if cfg.MetricsAddr != ":9090" {
		t.Errorf("MetricsAddr: want :9090, got %s", cfg.MetricsAddr)
	}
	if cfg.ReadTimeout != 30*time.Second {
		t.Errorf("ReadTimeout: want 30s, got %v", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != 30*time.Second {
		t.Errorf("WriteTimeout: want 30s, got %v", cfg.WriteTimeout)
	}
	if cfg.TargetURL != "http://backend:8081" {
		t.Errorf("TargetURL: want http://backend:8081, got %s", cfg.TargetURL)
	}
}

func TestFromEnv_CustomValues(t *testing.T) {
	t.Setenv("PROXY_TARGET_URL", "http://svc:3000")
	t.Setenv("PROXY_LISTEN_ADDR", ":9000")
	t.Setenv("PROXY_METRICS_ADDR", ":2112")
	t.Setenv("PROXY_READ_TIMEOUT_S", "10")
	t.Setenv("PROXY_WRITE_TIMEOUT_S", "15")

	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ListenAddr != ":9000" {
		t.Errorf("ListenAddr: want :9000, got %s", cfg.ListenAddr)
	}
	if cfg.ReadTimeout != 10*time.Second {
		t.Errorf("ReadTimeout: want 10s, got %v", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != 15*time.Second {
		t.Errorf("WriteTimeout: want 15s, got %v", cfg.WriteTimeout)
	}
}

func TestFromEnv_InvalidTimeout(t *testing.T) {
	t.Setenv("PROXY_TARGET_URL", "http://svc:3000")
	t.Setenv("PROXY_READ_TIMEOUT_S", "notanumber")

	_, err := FromEnv()
	if err == nil {
		t.Fatal("expected error for non-integer timeout, got nil")
	}
}

func TestFromEnv_NonPositiveTimeout(t *testing.T) {
	t.Setenv("PROXY_TARGET_URL", "http://svc:3000")
	t.Setenv("PROXY_WRITE_TIMEOUT_S", "0")

	_, err := FromEnv()
	if err == nil {
		t.Fatal("expected error for zero timeout, got nil")
	}
}
