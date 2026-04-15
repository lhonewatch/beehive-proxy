package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/example/beehive-proxy/internal/middleware"
)

// Config holds all runtime configuration for beehive-proxy.
type Config struct {
	TargetURL       string
	ListenAddr      string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	RequestTimeout  time.Duration
	MaxRetries      int
	RateLimit       int
	RateLimitWindow time.Duration
	CBThreshold     int
	CBCooldown      time.Duration
	CacheTTL        time.Duration
	AllowedOrigins  []string
	// IP filter
	IPFilterMode    string   // "allowlist", "blocklist", or ""
	IPFilterCIDRs   []string
}

// FromEnv builds a Config from environment variables.
func FromEnv() (*Config, error) {
	target := envString("BEEHIVE_TARGET_URL", "")
	if target == "" {
		return nil, errors.New("BEEHIVE_TARGET_URL is required")
	}
	if _, err := url.ParseRequestURI(target); err != nil {
		return nil, fmt.Errorf("invalid BEEHIVE_TARGET_URL: %w", err)
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
	requestTimeout, err := envDuration("BEEHIVE_REQUEST_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("BEEHIVE_REQUEST_TIMEOUT: %w", err)
	}
	rateLimitWindow, err := envDuration("BEEHIVE_RATE_LIMIT_WINDOW", time.Minute)
	if err != nil {
		return nil, fmt.Errorf("BEEHIVE_RATE_LIMIT_WINDOW: %w", err)
	}
	cbCooldown, err := envDuration("BEEHIVE_CB_COOLDOWN", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("BEEHIVE_CB_COOLDOWN: %w", err)
	}
	cacheTTL, err := envDuration("BEEHIVE_CACHE_TTL", 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("BEEHIVE_CACHE_TTL: %w", err)
	}

	ipMode := strings.ToLower(envString("BEEHIVE_IP_FILTER_MODE", ""))
	if ipMode != "" && ipMode != "allowlist" && ipMode != "blocklist" {
		return nil, fmt.Errorf("BEEHIVE_IP_FILTER_MODE must be 'allowlist', 'blocklist', or empty; got %q", ipMode)
	}
	var ipCIDRs []string
	if raw := envString("BEEHIVE_IP_FILTER_CIDRS", ""); raw != "" {
		for _, c := range strings.Split(raw, ",") {
			if t := strings.TrimSpace(c); t != "" {
				ipCIDRs = append(ipCIDRs, t)
			}
		}
	}

	return &Config{
		TargetURL:       target,
		ListenAddr:      envString("BEEHIVE_LISTEN_ADDR", ":8080"),
		ReadTimeout:     readTimeout,
		WriteTimeout:    writeTimeout,
		IdleTimeout:     idleTimeout,
		RequestTimeout:  requestTimeout,
		MaxRetries:      envInt("BEEHIVE_MAX_RETRIES", 3),
		RateLimit:       envInt("BEEHIVE_RATE_LIMIT", 100),
		RateLimitWindow: rateLimitWindow,
		CBThreshold:     envInt("BEEHIVE_CB_THRESHOLD", 5),
		CBCooldown:      cbCooldown,
		CacheTTL:        cacheTTL,
		AllowedOrigins:  strings.Split(envString("BEEHIVE_ALLOWED_ORIGINS", "*"), ","),
		IPFilterMode:    ipMode,
		IPFilterCIDRs:   ipCIDRs,
	}, nil
}

// IPFilterMiddleware returns the configured IP filter middleware, or nil if disabled.
func (c *Config) IPFilterMiddleware() func(http.Handler) http.Handler {
	if c.IPFilterMode == "" {
		return nil
	}
	mode := middleware.Allowlist
	if c.IPFilterMode == "blocklist" {
		mode = middleware.Blocklist
	}
	return middleware.NewIPFilter(mode, c.IPFilterCIDRs)
}

func envString(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func envDuration(key string, def time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	return time.ParseDuration(v)
}
