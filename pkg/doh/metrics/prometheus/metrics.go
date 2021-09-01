// Package prometheus defines a Prometheus metric implementation for DoH.
package prometheus

import (
	"errors"
	"fmt"

	dotmetrics "github.com/qdm12/dns/pkg/dot/metrics"
	middleware "github.com/qdm12/dns/pkg/middlewares/metrics"
	prom "github.com/qdm12/dns/pkg/prometheus"
)

type (
	dotDialMetrics      = dotmetrics.DialMetrics
	middlewareInterface = middleware.Interface
)

type Metrics struct {
	*counters
	dotDialMetrics
	middlewareInterface
}

var (
	ErrNewCounters = errors.New("failed creating counters metrics")
)

func New(settings prom.Settings,
	dotDialMetrics dotmetrics.DialMetrics,
	middlewareMetrics middleware.Interface,
) (metrics *Metrics, err error) {
	metrics = new(Metrics)

	metrics.counters, err = newCounters(settings)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNewCounters, err)
	}

	metrics.dotDialMetrics = dotDialMetrics
	metrics.middlewareInterface = middlewareMetrics

	return metrics, nil
}
