// Package prometheus defines shared elements for Prometheus.
package prometheus

import "github.com/prometheus/client_golang/prometheus"

type Settings struct {
	// Prefix, aka Subsystem, is the prefix string in front
	// of all metric names.
	Prefix string
	// Registry is the Prometheus registerer to use for the metrics.
	// It defaults to prometheus.DefaultRegisterer if unset.
	Registry prometheus.Registerer
}

func (s *Settings) SetDefaults() {
	if s.Registry == nil {
		s.Registry = prometheus.DefaultRegisterer
	}
}
