package prometheus

import (
	dot "github.com/qdm12/dns/v2/pkg/dot/metrics"
	dotnoop "github.com/qdm12/dns/v2/pkg/dot/metrics/noop"
	prom "github.com/qdm12/dns/v2/pkg/metrics/prometheus"
)

type Settings struct {
	// Prometheus defines common Prometheus settings.
	Prometheus prom.Settings
	// DoTDialMetrics is the metrics interface for the
	// DoT dialer. It defaults to a No-op implementation.
	DoTDialMetrics dot.DialMetrics
}

func (s *Settings) SetDefaults() {
	s.Prometheus.SetDefaults()

	if s.DoTDialMetrics == nil {
		s.DoTDialMetrics = dotnoop.New()
	}
}

func (s Settings) Validate() (err error) {
	return s.Prometheus.Validate()
}
