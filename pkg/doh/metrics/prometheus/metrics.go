// Package prometheus defines a Prometheus metric implementation for DoH.
package prometheus

import (
	"fmt"

	dotmetrics "github.com/qdm12/dns/v2/pkg/dot/metrics"
	middleware "github.com/qdm12/dns/v2/pkg/middlewares/metrics"
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

func New(settings Settings) (metrics *Metrics, err error) {
	settings.SetDefaults()

	metrics = new(Metrics)

	metrics.counters, err = newCounters(settings.Prometheus)
	if err != nil {
		return nil, fmt.Errorf("creating counters: %w", err)
	}

	metrics.dotDialMetrics = settings.DoTDialMetrics
	metrics.middlewareInterface = settings.MiddlewareMetrics

	return metrics, nil
}
