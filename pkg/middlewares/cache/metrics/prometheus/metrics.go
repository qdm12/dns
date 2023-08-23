// Package prometheus defines a Prometheus metric implementation
// for the cache.
package prometheus

import (
	"fmt"
)

type Metrics struct {
	*counters
	*gauges
	*labels
}

func New(settings Settings) (metrics *Metrics, err error) {
	settings.SetDefaults()

	metrics = new(Metrics)

	metrics.counters, err = newCounters(settings.Prometheus)
	if err != nil {
		return nil, fmt.Errorf("creating counters: %w", err)
	}

	metrics.gauges, err = newGauges(settings.Prometheus)
	if err != nil {
		return nil, fmt.Errorf("creating gauges: %w", err)
	}

	metrics.labels, err = newLabels(settings.Prometheus)
	if err != nil {
		return nil, fmt.Errorf("creating labels: %w", err)
	}

	return metrics, nil
}
