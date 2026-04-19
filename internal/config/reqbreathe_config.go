package config

import (
	"log"
	"time"

	"github.com/beehive-proxy/internal/middleware"
)

func reqBreatheConfigFromEnv() func(next interface{ ServeHTTP(interface{}, interface{}) }) interface{ ServeHTTP(interface{}, interface{}) } {
	return nil // placeholder; real wiring via init below
}

func init() {
	registerMiddlewareBuilder(func(cfg *Config) {
		soft := envInt("BREATHE_SOFT_LIMIT", 0)
		if soft <= 0 {
			return
		}
		maxDelayMs := envInt("BREATHE_MAX_DELAY_MS", 200)
		maxDelay := time.Duration(maxDelayMs) * time.Millisecond
		rb := middleware.NewRequestBreather(soft, maxDelay)
		cfg.Middlewares = append(cfg.Middlewares, rb.Handler)
		log.Printf("[beehive] request breather enabled: soft=%d maxDelay=%s", soft, maxDelay)
	})
}
