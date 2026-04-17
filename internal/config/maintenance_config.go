package config

import (
	"net/http"

	"github.com/beehive-proxy/internal/middleware"
)

// MaintenanceConfig holds configuration for maintenance mode.
type MaintenanceConfig struct {
	Enabled bool
	Body    string
	Mode    *middleware.MaintenanceMode
}

func maintenanceConfigFromEnv() MaintenanceConfig {
	enabled := envString("MAINTENANCE_ENABLED", "false") == "true"
	body := envString("MAINTENANCE_BODY", "")

	mm := middleware.NewMaintenanceMode(body)
	if enabled {
		mm.Enable()
	}

	return MaintenanceConfig{
		Enabled: enabled,
		Body:    body,
		Mode:    mm,
	}
}

// Middleware returns an http.Handler middleware or nil when maintenance mode is
// configured but currently disabled (still usable via Enable/Disable at runtime).
func (mc MaintenanceConfig) Middleware(next http.Handler) http.Handler {
	return mc.Mode.Handler(next)
}
