package config

import (
	"os"
	"testing"
	"time"
)

func setEnv(t *testing.T, pairs ...string) {
	t.Helper()
	for i := 0; i < len(pairs)-1; i += 2 {
		t.Setenv(pairs[i], pairs[i+1])
	}
}

func TestFromEnv_MissingTargetURL(t *testing.T) {
	os.Unsetenv("TARGET_URL")
	_, err := FromEnv()
	if err == nil {
		t.Fatal("expected error for missing TARGET_URL")
	}
}

func TestFromEnv_Defaults(t *testing.T) {
	setEnv(t, "TARGET_URL", "http://backend")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ListenAddr != ":8080" {
		t.Errorf("expected :8080, got %s", cfg.ListenAddr)
	}
	if cfg.ReadTimeout != 30000*time.Millisecond {
		t.Errorf("unexpected ReadTimeout: %v", cfg.ReadTimeout)
	}
	if cfg.CBThreshold != 5 {
		t.Errorf("expected CB_THRESHOLD default 5, got %d", cfg.CBThreshold)
	}
	if cfg.CBCooldown != 10000*time.Millisecond {
		t.Errorf("expected CB_COOLDOWN default 10s, got %v", cfg.CBCooldown)
	}
}

func TestFromEnv_CustomValues(t *testing.T) {
	setEnv(t,
		"TARGET_URL", "http://custom",
		"LISTEN_ADDR", ":9090",
		"CB_THRESHOLD", "10",
		"CB_COOLDOWN_MS", "5000",
	)
	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ListenAddr != ":9090" {
		t.Errorf("expected :9090, got %s", cfg.ListenAddr)
	}
	if cfg.CBThreshold != 10 {
		t.Errorf("expected 10, got %d", cfg.CBThreshold)
	}
	if cfg.CBCooldown != 5*time.Second {
		t.Errorf("expected 5s, got %v", cfg.CBCooldown)
	}
}

func TestFromEnv_InvalidTimeout(t *testing.T) {
	setEnv(t, "TARGET_URL", "http://backend", "READ_TIMEOUT_MS", "notanint")
	_, err := FromEnv()
	if err == nil {
		t.Fatal("expected error for invalid READ_TIMEOUT_MS")
	}
}

func TestFromEnv_NonPositiveTimeout(t *testing.T) {
	setEnv(t, "TARGET_URL", "http://backend", "READ_TIMEOUT_MS", "0")
	_, err := FromEnv()
	if err == nil {
		t.Fatal("expected error for zero READ_TIMEOUT_MS")
	}
}

func TestFromEnv_InvalidCBThreshold(t *testing.T) {
	setEnv(t, "TARGET_URL", "http://backend", "CB_THRESHOLD", "bad")
	_, err := FromEnv()
	if err == nil {
		t.Fatal("expected error for invalid CB_THRESHOLD")
	}
}
