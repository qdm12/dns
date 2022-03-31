package helpers

import (
	"github.com/prometheus/client_golang/prometheus"
)

func NewCounter(prefix, name, help string) prometheus.Counter { //nolint:ireturn
	opts := prometheus.CounterOpts(newOpts(prefix, name, help))
	return prometheus.NewCounter(opts)
}

func NewCounterVec(prefix, name, help string, labelNames []string) *prometheus.CounterVec {
	opts := prometheus.CounterOpts(newOpts(prefix, name, help))
	return prometheus.NewCounterVec(opts, labelNames)
}
