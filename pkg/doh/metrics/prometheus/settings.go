package prometheus

import (
	dotnoop "github.com/qdm12/dns/v2/pkg/dot/metrics/noop"
	prom "github.com/qdm12/dns/v2/pkg/metrics/prometheus"
	"github.com/qdm12/gosettings"
)

type Settings struct {
	// Prometheus defines common Prometheus settings.
	Prometheus prom.Settings
	// DoTDialMetrics is the metrics interface for the
	// DoT dialer. It defaults to a No-op implementation.
	DoTDialMetrics DialMetrics
}

func (s *Settings) SetDefaults() {
	s.Prometheus.SetDefaults()
	s.DoTDialMetrics = gosettings.DefaultComparable[DialMetrics](s.DoTDialMetrics, dotnoop.New())
}

func (s Settings) Validate() (err error) {
	return s.Prometheus.Validate()
}
