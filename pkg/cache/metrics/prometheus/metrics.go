// Package prometheus defines a Prometheus metric implementation
// for the cache.
package prometheus

import (
	"errors"
	"fmt"
)

type Metrics struct {
	*counters
	*gauges
	*labels
}

var (
	ErrNewCounters = errors.New("failed creating counters metrics")
	ErrNewGauges   = errors.New("failed creating gauges metrics")
	ErrNewLabels   = errors.New("failed creating labels metrics")
)

func New(settings Settings) (metrics *Metrics, err error) {
	settings.setDefaults()

	metrics = new(Metrics)

	metrics.counters, err = newCounters(settings.Prometheus)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNewCounters, err)
	}

	metrics.gauges, err = newGauges(settings.Prometheus)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNewGauges, err)
	}

	metrics.labels, err = newLabels(settings.Prometheus)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNewLabels, err)
	}

	return metrics, nil
}
