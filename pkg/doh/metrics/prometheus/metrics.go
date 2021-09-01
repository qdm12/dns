// Package prometheus defines a Prometheus metric implementation for DoH.
package prometheus

import (
	"errors"
	"fmt"

	dotmetrics "github.com/qdm12/dns/pkg/dot/metrics"
	middleware "github.com/qdm12/dns/pkg/middlewares/metrics"
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

func New(settings Settings) (metrics *Metrics, err error) {
	settings.setDefaults()

	metrics = new(Metrics)

	metrics.counters, err = newCounters(settings.Prometheus)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNewCounters, err)
	}

	metrics.dotDialMetrics = settings.DoTDialMetrics
	metrics.middlewareInterface = settings.MiddlewareMetrics

	return metrics, nil
}
