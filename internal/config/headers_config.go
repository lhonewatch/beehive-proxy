package config

import (
	"strings"

	"github.com/beehive-proxy/internal/middleware"
)

// HeadersConfig holds header injection/removal settings sourced from env vars.
//
// Environment variables:
//   PROXY_ADD_REQUEST_HEADERS   — comma-separated key=value pairs
//   PROXY_REMOVE_REQUEST_HEADERS — comma-separated header names
//   PROXY_ADD_RESPONSE_HEADERS  — comma-separated key=value pairs
//   PROXY_REMOVE_RESPONSE_HEADERS — comma-separated header names
type HeadersConfig struct {
	AddRequest      map[string]string
	RemoveRequest   []string
	AddResponse     map[string]string
	RemoveResponse  []string
}

func parseKVList(raw string) map[string]string {
	out := make(map[string]string)
	if raw == "" {
		return out
	}
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			out[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return out
}

func parseNameList(raw string) []string {
	if raw == "" {
		return nil
	}
	var out []string
	for _, s := range strings.Split(raw, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// headersConfigFromEnv reads header configuration from environment variables.
func headersConfigFromEnv() HeadersConfig {
	return HeadersConfig{
		AddRequest:     parseKVList(envString("PROXY_ADD_REQUEST_HEADERS", "")),
		RemoveRequest:  parseNameList(envString("PROXY_REMOVE_REQUEST_HEADERS", "")),
		AddResponse:    parseKVList(envString("PROXY_ADD_RESPONSE_HEADERS", "")),
		RemoveResponse: parseNameList(envString("PROXY_REMOVE_RESPONSE_HEADERS", "")),
	}
}

// Middleware builds a middleware.HeadersOptions from this config.
func (h HeadersConfig) Middleware() middleware.HeadersOptions {
	return middleware.HeadersOptions{
		RequestHeaders:        h.AddRequest,
		RemoveRequestHeaders:  h.RemoveRequest,
		ResponseHeaders:       h.AddResponse,
		RemoveResponseHeaders: h.RemoveResponse,
	}
}
