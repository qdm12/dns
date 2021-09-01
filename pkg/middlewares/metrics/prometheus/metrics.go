// Package prometheus defines a Prometheus metric implementation
// for a DNS server middleware.
package prometheus

import (
	"errors"
	"fmt"

	prom "github.com/qdm12/dns/pkg/prometheus"
)

type Metrics struct {
	*counters
	*gauges
}

var (
	ErrNewCounters = errors.New("failed creating metrics counters")
	ErrNewGauges   = errors.New("failed creating metrics gauges")
)

func New(settings prom.Settings) (
	metrics *Metrics, err error) {
	metrics = new(Metrics)

	metrics.counters, err = newCounters(settings)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNewCounters, err)
	}

	metrics.gauges, err = newGauges(settings)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNewGauges, err)
	}

	return metrics, nil
}
