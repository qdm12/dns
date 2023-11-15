// Package prometheus defines a Prometheus metric implementation for DoT.
package prometheus

import (
	"fmt"
)

type Metrics struct {
	*counters
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

	return metrics, nil
}
