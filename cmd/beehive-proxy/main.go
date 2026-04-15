package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"beehive-proxy/internal/config"
	"beehive-proxy/internal/metrics"
	"beehive-proxy/internal/middleware"
	"beehive-proxy/internal/proxy"
	"beehive-proxy/internal/tracing"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.FromEnv()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	retryTransport := middleware.NewRetryTransport(nil, middleware.RetryConfig{
		MaxAttempts: cfg.RetryMaxAttempts,
		Delay:       cfg.RetryDelay,
		Logger:      logger,
	})

	proxyHandler := proxy.NewHandler(cfg.TargetURL, tracing.DefaultTracerFunc, metrics.ObserveRequest, retryTransport)

	chain := middleware.Recovery(
		middleware.RequestLogger(logger,
			middleware.NewCircuitBreaker(middleware.CircuitBreakerConfig{
				Threshold: cfg.CBThreshold,
				Cooldown:  cfg.CBCooldown,
			},
				middleware.NewRateLimiter(cfg.RateLimit, cfg.RateLimitWindow, proxyHandler),
			),
		),
	)

	mux := http.NewServeMux()
	mux.Handle(cfg.MetricsPath, promhttp.Handler())
	mux.Handle("/", chain)

	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		logger.Info("starting beehive-proxy", "addr", cfg.ListenAddr, "target", cfg.TargetURL)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	logger.Info("shutting down gracefully")
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", "error", err)
	}
}
