package prometheus

import (
	prom "github.com/qdm12/dns/pkg/prometheus"
)

type Settings struct {
	// Prometheus defines common Prometheus settings.
	Prometheus prom.Settings
}

func (s *Settings) setDefaults() {
	s.Prometheus.SetDefaults()
}
