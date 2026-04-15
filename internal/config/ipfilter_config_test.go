package config

import (
	"os"
	"testing"
)

func TestFromEnv_IPFilterDefaults(t *testing.T) {
	setEnv(t, "BEEHIVE_TARGET_URL", "http://backend:8080")

	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.IPFilterMode != "" {
		t.Errorf("expected empty IPFilterMode, got %q", cfg.IPFilterMode)
	}
	if len(cfg.IPFilterCIDRs) != 0 {
		t.Errorf("expected no CIDRs, got %v", cfg.IPFilterCIDRs)
	}
}

func TestFromEnv_IPFilterAllowlist(t *testing.T) {
	setEnv(t, "BEEHIVE_TARGET_URL", "http://backend:8080")
	setEnv(t, "BEEHIVE_IP_FILTER_MODE", "allowlist")
	setEnv(t, "BEEHIVE_IP_FILTER_CIDRS", "10.0.0.0/8, 192.168.1.0/24")

	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.IPFilterMode != "allowlist" {
		t.Errorf("expected allowlist, got %q", cfg.IPFilterMode)
	}
	if len(cfg.IPFilterCIDRs) != 2 {
		t.Errorf("expected 2 CIDRs, got %d", len(cfg.IPFilterCIDRs))
	}
}

func TestFromEnv_IPFilterBlocklist(t *testing.T) {
	setEnv(t, "BEEHIVE_TARGET_URL", "http://backend:8080")
	setEnv(t, "BEEHIVE_IP_FILTER_MODE", "blocklist")
	setEnv(t, "BEEHIVE_IP_FILTER_CIDRS", "203.0.113.0/24")

	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.IPFilterMode != "blocklist" {
		t.Errorf("expected blocklist, got %q", cfg.IPFilterMode)
	}
}

func TestFromEnv_IPFilterInvalidMode(t *testing.T) {
	setEnv(t, "BEEHIVE_TARGET_URL", "http://backend:8080")
	setEnv(t, "BEEHIVE_IP_FILTER_MODE", "whitelist")

	_, err := FromEnv()
	if err == nil {
		t.Fatal("expected error for invalid IP filter mode")
	}
}

func TestFromEnv_IPFilterMiddlewareNilWhenDisabled(t *testing.T) {
	setEnv(t, "BEEHIVE_TARGET_URL", "http://backend:8080")
	os.Unsetenv("BEEHIVE_IP_FILTER_MODE")

	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.IPFilterMiddleware() != nil {
		t.Error("expected nil middleware when IP filter is disabled")
	}
}

func TestFromEnv_IPFilterMiddlewareNonNilWhenEnabled(t *testing.T) {
	setEnv(t, "BEEHIVE_TARGET_URL", "http://backend:8080")
	setEnv(t, "BEEHIVE_IP_FILTER_MODE", "allowlist")
	setEnv(t, "BEEHIVE_IP_FILTER_CIDRS", "127.0.0.0/8")

	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.IPFilterMiddleware() == nil {
		t.Error("expected non-nil middleware when IP filter is enabled")
	}
}
