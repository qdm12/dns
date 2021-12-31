// Package prometheus defines a Prometheus metric implementation
// for a DNS server middleware.
package prometheus

import (
	"errors"
	"fmt"
)

type Metrics struct {
	*counters
	*gauges
}

var (
	ErrNewCounters = errors.New("failed creating metrics counters")
	ErrNewGauges   = errors.New("failed creating metrics gauges")
)

func New(settings Settings) (metrics *Metrics, err error) {
	settings.SetDefaults()

	metrics = new(Metrics)

	metrics.counters, err = newCounters(settings.Prometheus)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNewCounters, err)
	}

	metrics.gauges, err = newGauges(settings.Prometheus)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNewGauges, err)
	}

	return metrics, nil
}
