package helpers

import (
	"github.com/prometheus/client_golang/prometheus"
)

func Register(registry prometheus.Registerer,
	collectors ...prometheus.Collector) (err error) {
	for _, collector := range collectors {
		_ = registry.Unregister(collector)

		if err = registry.Register(collector); err != nil {
			return err
		}
	}
	return nil
}
