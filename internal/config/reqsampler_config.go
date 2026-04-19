package config

import (
	"log"
	"net/http"
	"strconv"

	"github.com/example/beehive-proxy/internal/middleware"
)

func reqSamplerMiddlewareFromEnv() func(http.Handler) http.Handler {
	raw := envString("SAMPLER_RATE", "")
	if raw == "" {
		return nil
	}
	rate, err := strconv.ParseFloat(raw, 64)
	if err != nil || rate < 0 || rate > 1 {
		log.Printf("[config] invalid SAMPLER_RATE %q, disabling sampler", raw)
		return nil
	}
	log.Printf("[config] request sampler enabled at rate=%.2f", rate)
	return middleware.NewRequestSampler(rate)
}

func init() {
	registerMiddlewareFactory(reqSamplerMiddlewareFromEnv)
}
