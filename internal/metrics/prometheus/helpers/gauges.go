package helpers

import (
	"github.com/prometheus/client_golang/prometheus"
)

func NewGauge(prefix, name, help string) prometheus.Gauge { //nolint:ireturn
	opts := prometheus.GaugeOpts(newOpts(prefix, name, help))
	return prometheus.NewGauge(opts)
}

func NewGaugeVec(prefix, name, help string, labelNames []string) *prometheus.GaugeVec {
	opts := prometheus.GaugeOpts(newOpts(prefix, name, help))
	return prometheus.NewGaugeVec(opts, labelNames)
}
