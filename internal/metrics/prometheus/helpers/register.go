package helpers

import (
	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusRegistry interface {
	prometheus.Registerer
}

type PrometheusCollector interface {
	prometheus.Collector
}

func Register(registry PrometheusRegistry,
	collectors ...PrometheusCollector) (err error) {
	for _, collector := range collectors {
		_ = registry.Unregister(collector)

		if err = registry.Register(collector); err != nil {
			return err
		}
	}
	return nil
}
