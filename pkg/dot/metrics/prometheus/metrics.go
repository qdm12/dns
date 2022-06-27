// Package prometheus defines a Prometheus metric implementation for DoT.
package prometheus

import (
	"fmt"

	middleware "github.com/qdm12/dns/v2/pkg/middlewares/metrics"
)

type middlewareInterface = middleware.Interface

type Metrics struct {
	*counters
	middlewareInterface
}

func New(settings Settings) (metrics *Metrics, err error) {
	settings.SetDefaults()

	metrics = new(Metrics)

	metrics.counters, err = newCounters(settings.Prometheus)
	if err != nil {
		return nil, fmt.Errorf("creating counters: %w", err)
	}

	metrics.middlewareInterface = settings.MiddlewareMetrics

	return metrics, nil
}
