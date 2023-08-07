// Package prometheus defines a Prometheus metric implementation for DoH.
package prometheus

import (
	"fmt"
)

type (
	// unexported alias so it is not exposed through
	// the Metrics struct.
	dotDialMetrics = DialMetrics
)

type Metrics struct {
	*counters
	dotDialMetrics
}

func New(settings Settings) (metrics *Metrics, err error) {
	settings.SetDefaults()

	metrics = new(Metrics)

	metrics.counters, err = newCounters(settings.Prometheus)
	if err != nil {
		return nil, fmt.Errorf("creating counters: %w", err)
	}

	metrics.dotDialMetrics = settings.DoTDialMetrics

	return metrics, nil
}
