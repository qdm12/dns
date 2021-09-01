// Package prometheus defines a Prometheus metric implementation for DoT.
package prometheus

import (
	"errors"
	"fmt"

	middleware "github.com/qdm12/dns/pkg/middlewares/metrics"
	prom "github.com/qdm12/dns/pkg/prometheus"
)

type middlewareInterface = middleware.Interface

type Metrics struct {
	*counters
	middlewareInterface
}

var (
	ErrNewCounters = errors.New("failed creating metrics counters")
)

func New(settings prom.Settings,
	middlewareMetrics middleware.Interface) (
	metrics *Metrics, err error) {
	metrics = new(Metrics)

	metrics.counters, err = newCounters(settings)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNewCounters, err)
	}

	metrics.middlewareInterface = middlewareMetrics

	return metrics, nil
}
