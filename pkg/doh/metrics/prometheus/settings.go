package prometheus

import (
	dot "github.com/qdm12/dns/pkg/dot/metrics"
	dotnoop "github.com/qdm12/dns/pkg/dot/metrics/noop"
	prom "github.com/qdm12/dns/pkg/metrics/prometheus"
	middleware "github.com/qdm12/dns/pkg/middlewares/metrics"
	middlewarenoop "github.com/qdm12/dns/pkg/middlewares/metrics/noop"
)

type Settings struct {
	// Prometheus defines common Prometheus settings.
	Prometheus prom.Settings
	// DoTDialMetrics is the metrics interface for the
	// DoT dialer. It defaults to a No-op implementation.
	DoTDialMetrics dot.DialMetrics
	// MiddlewareMetrics is the metrics interface for the
	// DNS middleware. It defaults to a No-op implementation.
	MiddlewareMetrics middleware.Interface
}

func (s *Settings) SetDefaults() {
	s.Prometheus.SetDefaults()

	if s.DoTDialMetrics == nil {
		s.DoTDialMetrics = dotnoop.New()
	}

	if s.MiddlewareMetrics == nil {
		s.MiddlewareMetrics = middlewarenoop.New()
	}
}

func (s Settings) Validate() (err error) {
	return s.Prometheus.Validate()
}
