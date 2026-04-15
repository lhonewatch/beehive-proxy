package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/example/beehive-proxy/internal/config"
	"github.com/example/beehive-proxy/internal/metrics"
	"github.com/example/beehive-proxy/internal/middleware"
	"github.com/example/beehive-proxy/internal/proxy"
	"github.com/example/beehive-proxy/internal/tracing"
)

func main() {
	logger := log.New(os.Stdout, "", 0)

	cfg, err := config.FromEnv()
	if err != nil {
		logger.Fatalf("config error: %v", err)
	}

	transport := middleware.NewRetryTransport(http.DefaultTransport, cfg.MaxRetries)

	proxyHandler := proxy.NewHandler(cfg.TargetURL, tracing.DefaultTracerFunc, transport)

	chain := proxyHandler

	// IP filter (optional)
	if ipf := cfg.IPFilterMiddleware(); ipf != nil {
		chain = ipf(chain)
	}

	chain = middleware.NewCacheMiddleware(middleware.NewResponseCache(cfg.CacheTTL))(chain)
	chain = middleware.NewCompress()(chain)
	chain = middleware.NewCORS(middleware.DefaultCORSOptions(cfg.AllowedOrigins...))(chain)
	chain = middleware.NewCircuitBreaker(cfg.CBThreshold, cfg.CBCooldown)(chain)
	chain = middleware.NewRateLimiter(cfg.RateLimit, cfg.RateLimitWindow)(chain)
	chain = middleware.NewTimeout(cfg.RequestTimeout)(chain)
	chain = middleware.RequestLogger(logger)(chain)
	chain = middleware.Recovery(logger)(chain)

	hc := middleware.NewHealthChecker(chain)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/", metrics.Middleware(hc))

	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	logger.Printf(`{"level":"info","msg":"starting beehive-proxy","addr":%q,"target":%q,"ip_filter":%q}`,
		cfg.ListenAddr, cfg.TargetURL, cfg.IPFilterMode)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("server error: %v", err)
	}
}
