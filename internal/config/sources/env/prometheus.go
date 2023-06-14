package env

import (
	"github.com/qdm12/dns/v2/internal/config/settings"
)

func (r *Reader) readPrometheus() (settings settings.Prometheus) {
	settings.ListeningAddress = r.env.String("METRICS_PROMETHEUS_ADDRESS")
	settings.Subsystem = r.env.Get("METRICS_PROMETHEUS_SUBSYSTEM")
	return settings
}
