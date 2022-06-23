package env

import (
	"os"

	"github.com/qdm12/dns/v2/internal/config/settings"
)

func readPrometheus() (settings settings.Prometheus) {
	settings.ListeningAddress = os.Getenv("METRICS_PROMETHEUS_ADDRESS")
	settings.Subsystem = envToStringPtr("METRICS_PROMETHEUS_SUBSYSTEM")
	return settings
}
