package helpers

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var cache = metricCache{ //nolint:gochecknoglobals
	nameToCounter:    map[string]prometheus.Counter{},
	nameToCounterVec: map[string]*prometheus.CounterVec{},
	nameToGauge:      map[string]prometheus.Gauge{},
	nameToGaugeVec:   map[string]*prometheus.GaugeVec{},
}

type metricCache struct {
	nameToCounter    map[string]prometheus.Counter
	nameToCounterVec map[string]*prometheus.CounterVec
	nameToGauge      map[string]prometheus.Gauge
	nameToGaugeVec   map[string]*prometheus.GaugeVec
	mutex            sync.Mutex
}

func (m *metricCache) getCounter(prefix, name string) prometheus.Counter {
	return getFromMetricCache(prefix, name, &m.mutex, m.nameToCounter)
}

func (m *metricCache) setCounter(prefix, name string, counter prometheus.Counter) {
	setToMetricCache(prefix, name, &m.mutex, m.nameToCounter, counter)
}

func (m *metricCache) getCounterVec(prefix, name string) *prometheus.CounterVec {
	return getFromMetricCache(prefix, name, &m.mutex, m.nameToCounterVec)
}

func (m *metricCache) setCounterVec(prefix, name string, counterVec *prometheus.CounterVec) {
	setToMetricCache(prefix, name, &m.mutex, m.nameToCounterVec, counterVec)
}

func (m *metricCache) getGauge(prefix, name string) prometheus.Gauge {
	return getFromMetricCache(prefix, name, &m.mutex, m.nameToGauge)
}

func (m *metricCache) setGauge(prefix, name string, gauge prometheus.Gauge) {
	setToMetricCache(prefix, name, &m.mutex, m.nameToGauge, gauge)
}

func (m *metricCache) getGaugeVec(prefix, name string) *prometheus.GaugeVec {
	return getFromMetricCache(prefix, name, &m.mutex, m.nameToGaugeVec)
}

func (m *metricCache) setGaugeVec(prefix, name string, gaugeVec *prometheus.GaugeVec) {
	setToMetricCache(prefix, name, &m.mutex, m.nameToGaugeVec, gaugeVec)
}

func getFromMetricCache[T any](prefix, name string, mutex *sync.Mutex, //nolint:ireturn
	nameToCollector map[string]T,
) (collector T) {
	mutex.Lock()
	defer mutex.Unlock()
	return nameToCollector[prefix+"_"+name]
}

func setToMetricCache[T any](prefix, name string, mutex *sync.Mutex,
	nameToCollector map[string]T, collector T,
) {
	mutex.Lock()
	defer mutex.Unlock()
	nameToCollector[prefix+"_"+name] = collector
}
