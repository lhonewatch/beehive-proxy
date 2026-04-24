package config

// This file ensures the reqburst init() function is linked when the config
// package is imported, registering the burst middleware builder.
// The actual registration is performed in reqburst_config.go init().
