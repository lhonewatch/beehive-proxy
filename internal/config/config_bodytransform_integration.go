package config

// init registers the BodyTransform configuration extension, which populates
// the BodyTransform field from environment variables during config loading.
func init() {
	registerExtension(func(c *Config) error {
		c.BodyTransform = bodyTransformConfigFromEnv()
		return nil
	})
}
