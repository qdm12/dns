// Package prometheus defines a Prometheus metric implementation
// for the filter.
package prometheus

import (
	"fmt"
)

type Metrics struct {
	*counters
	*gauges
}

func New(settings Settings) (metrics *Metrics, err error) {
	settings.SetDefaults()

	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("settings validation: %w", err)
	}

	metrics = new(Metrics)

	metrics.counters, err = newCounters(settings.Prometheus)
	if err != nil {
		return nil, fmt.Errorf("creating counters: %w", err)
	}

	metrics.gauges, err = newGauges(settings.Prometheus)
	if err != nil {
		return nil, fmt.Errorf("creating gauges: %w", err)
	}

	return metrics, nil
}
