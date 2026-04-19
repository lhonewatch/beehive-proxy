package config

// Throttle fields surfaced on Config for inspection in tests and main.
func init() {
	registerConfigExtractor(func(c *Config) {
		tc := reqThrottleConfigFromEnv()
		c.ThrottleEnabled = tc.enabled
		c.ThrottleRate = tc.rate
		c.ThrottleBurst = tc.burst
	})
}
