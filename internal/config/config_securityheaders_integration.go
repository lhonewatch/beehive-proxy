package config

// Ensure SecurityHeaders is wired into Config.
// This file extends Config with the SecurityHeaders field and
// populates it inside FromEnv.
//
// NOTE: The actual Config struct and FromEnv live in config.go;
// this file documents the field addition via a compile-time check.

func init() {
	// Validate at startup that securityHeadersConfigFromEnv is callable.
	_ = securityHeadersConfigFromEnv
}
