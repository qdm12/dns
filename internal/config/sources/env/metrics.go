package env

import (
	"github.com/qdm12/dns/v2/internal/config/settings"
)

func (r *Reader) readMetrics() (settings settings.Metrics) {
	settings.Type = r.env.String("METRICS_TYPE")
	settings.Prometheus = r.readPrometheus()
	return settings
}
