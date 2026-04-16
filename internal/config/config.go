package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration for beehive-proxy.
type Config struct {
	TargetURL       *url.URL
	ListenAddr      string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	MetricsAddr     string
	MaxBodyBytes    int64
	RateLimit       RateLimitConfig
	CircuitBreaker  CircuitBreakerConfig
	Retry           RetryConfig
	CORS            CORSConfig
	Cache           CacheConfig
	IPFilter        IPFilterConfig
	Rewrite         RewriteConfig
	Headers         HeadersConfig
	BasicAuth       BasicAuthConfig
	JWT             JWTConfig
}

// FromEnv builds a Config from environment variables.
func FromEnv() (*Config, error) {
	rawURL := envString("TARGET_URL", "")
	if rawURL == "" {
		return nil, fmt.Errorf("TARGET_URL is required")
	}
	target, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid TARGET_URL: %w", err)
	}

	ipFilter, err := ipFilterConfigFromEnv()
	if err != nil {
		return nil, err
	}

	headers, err := headersConfigFromEnv()
	if err != nil {
		return nil, err
	}

	basicAuth, err := basicAuthConfigFromEnv()
	if err != nil {
		return nil, err
	}

	jwt, err := jwtConfigFromEnv()
	if err != nil {
		return nil, err
	}
	if err := jwt.Validate(); err != nil {
		return nil, err
	}

	return &Config{
		TargetURL:    target,
		ListenAddr:   envString("LISTEN_ADDR", ":8080"),
		MetricsAddr:  envString("METRICS_ADDR", ":9090"),
		ReadTimeout:  envDuration("READ_TIMEOUT", 30*time.Second),
		WriteTimeout: envDuration("WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:  envDuration("IDLE_TIMEOUT", 90*time.Second),
		MaxBodyBytes: int64(envInt("MAX_BODY_BYTES", 1<<20)),
		RateLimit: RateLimitConfig{
			RequestsPerSecond: envInt("RATE_LIMIT_RPS", 100),
			Burst:             envInt("RATE_LIMIT_BURST", 20),
		},
		CircuitBreaker: CircuitBreakerConfig{
			Threshold: envInt("CB_THRESHOLD", 5),
			Cooldown:  envDuration("CB_COOLDOWN", 10*time.Second),
		},
		Retry: RetryConfig{
			MaxAttempts: envInt("RETRY_MAX_ATTEMPTS", 3),
			Delay:       envDuration("RETRY_DELAY", 100*time.Millisecond),
		},
		CORS:      DefaultCORSConfig(),
		Cache:     CacheConfig{TTL: envDuration("CACHE_TTL", 0)},
		IPFilter:  ipFilter,
		Rewrite:   rewriteConfigFromEnv(),
		Headers:   headers,
		BasicAuth: basicAuth,
		JWT:       jwt,
	}, nil
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

func envDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
