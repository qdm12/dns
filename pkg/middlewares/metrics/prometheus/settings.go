package prometheus

import (
	prom "github.com/qdm12/dns/v2/pkg/metrics/prometheus"
)

type Settings struct {
	// Prometheus defines common Prometheus settings.
	Prometheus prom.Settings
}

func (s *Settings) SetDefaults() {
	s.Prometheus.SetDefaults()
}

func (s Settings) Validate() (err error) {
	return s.Prometheus.Validate()
}
