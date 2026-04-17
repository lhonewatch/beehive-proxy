package config

func init() {
	configLoaders = append(configLoaders, func(c *Config) error {
		c.Maintenance = maintenanceConfigFromEnv()
		return nil
	})
}
