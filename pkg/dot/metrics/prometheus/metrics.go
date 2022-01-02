// Package prometheus defines a Prometheus metric implementation for DoT.
package prometheus

import (
	"errors"
	"fmt"

	middleware "github.com/qdm12/dns/v2/pkg/middlewares/metrics"
)

type middlewareInterface = middleware.Interface

type Metrics struct {
	*counters
	middlewareInterface
}

var (
	ErrNewCounters = errors.New("failed creating metrics counters")
)

func New(settings Settings) (metrics *Metrics, err error) {
	settings.SetDefaults()

	metrics = new(Metrics)

	metrics.counters, err = newCounters(settings.Prometheus)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNewCounters, err)
	}

	metrics.middlewareInterface = settings.MiddlewareMetrics

	return metrics, nil
}
