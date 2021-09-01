package prometheus

import "github.com/prometheus/client_golang/prometheus"

type Settings struct {
	// Prefix, aka Subsystem, is the prefix string in front
	// of all metric names.
	Prefix   string
	Registry prometheus.Registerer
}
