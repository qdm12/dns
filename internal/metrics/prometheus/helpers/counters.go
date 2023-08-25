package helpers

import (
	"github.com/prometheus/client_golang/prometheus"
)

func NewCounter(prefix, name, help string) prometheus.Counter {
	counter := cache.getCounter(prefix, name)
	if counter != nil {
		return counter
	}
	opts := prometheus.CounterOpts(newOpts(prefix, name, help))
	counter = prometheus.NewCounter(opts)
	cache.setCounter(prefix, name, counter)
	return counter
}

func NewCounterVec(prefix, name, help string, labelNames []string) *prometheus.CounterVec {
	counterVec := cache.getCounterVec(prefix, name)
	if counterVec != nil {
		return counterVec
	}
	opts := prometheus.CounterOpts(newOpts(prefix, name, help))
	counterVec = prometheus.NewCounterVec(opts, labelNames)
	cache.setCounterVec(prefix, name, counterVec)
	return counterVec
}
