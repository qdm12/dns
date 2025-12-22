package helpers

import "github.com/prometheus/client_golang/prometheus"

func NewHistogramVec(prefix, name, help string, labelNames []string, buckets []float64) *prometheus.HistogramVec {
	histogramVec := cache.getHistogramVec(prefix, name)
	if histogramVec != nil {
		return histogramVec
	}
	opts := prometheus.HistogramOpts{
		Subsystem: prefix,
		Name:      name,
		Help:      help,
		Buckets:   buckets,
	}
	histogramVec = prometheus.NewHistogramVec(opts, labelNames)
	cache.setHistogramVec(prefix, name, histogramVec)
	return histogramVec
}
