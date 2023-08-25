package helpers

import (
	"github.com/prometheus/client_golang/prometheus"
)

func NewGauge(prefix, name, help string) prometheus.Gauge {
	gauge := cache.getGauge(prefix, name)
	if gauge != nil {
		return gauge
	}
	opts := prometheus.GaugeOpts(newOpts(prefix, name, help))
	gauge = prometheus.NewGauge(opts)
	cache.setGauge(prefix, name, gauge)
	return gauge
}

func NewGaugeVec(prefix, name, help string, labelNames []string) *prometheus.GaugeVec {
	gaugeVec := cache.getGaugeVec(prefix, name)
	if gaugeVec != nil {
		return gaugeVec
	}
	opts := prometheus.GaugeOpts(newOpts(prefix, name, help))
	gaugeVec = prometheus.NewGaugeVec(opts, labelNames)
	cache.setGaugeVec(prefix, name, gaugeVec)
	return gaugeVec
}
