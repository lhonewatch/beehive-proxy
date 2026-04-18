package config

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/andriikushch/beehive-proxy/internal/middleware"
)

func init() {
	registerMiddlewareBuilder(slowLogMiddlewareFromEnv)
}

func slowLogMiddlewareFromEnv(cfg *Config, logger *slog.Logger) (func(http.Handler) http.Handler, error) {
	enabled := envString("SLOW_LOG_ENABLED", "false")
	if enabled != "true" {
		return nil, nil
	}
	thresholdStr := envString("SLOW_LOG_THRESHOLD", "500ms")
	threshold, err := time.ParseDuration(thresholdStr)
	if err != nil {
		return nil, err
	}
	_ = cfg
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}
	return middleware.NewSlowLog(threshold, logger), nil
}
