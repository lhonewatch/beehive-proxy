package config_test

import (
	"net/http"
	"time"

	"github.com/beehive-proxy/internal/middleware"
)

// buildCloneMiddlewareForTest constructs the clone middleware directly,
// bypassing the env-driven init() registration used in production.
func buildCloneMiddlewareForTest(targetURL string) func(http.Handler) http.Handler {
	client := &http.Client{Timeout: 2 * time.Second}
	return middleware.NewRequestClone(targetURL, client)
}

// import_wait sleeps for the given number of milliseconds, giving
// background goroutines time to complete in tests.
func import_wait(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

// FromEnvForTest is an alias so config_test package can call FromEnv
// without importing it as an external package (same package tests).
func FromEnvForTest() (*Config, error) {
	return FromEnv()
}
