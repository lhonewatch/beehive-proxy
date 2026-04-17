package config

func init() {
	registerExtension(func(c *Config) error {
		c.BodyTransform = bodyTransformConfigFromEnv()
		return nil
	})
}
