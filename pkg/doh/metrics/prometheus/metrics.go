// Package prometheus defines a Prometheus metric implementation for DoH.
package prometheus

import (
	"fmt"

	dotmetrics "github.com/qdm12/dns/v2/pkg/dot/metrics"
)

type (
	dotDialMetrics = dotmetrics.DialMetrics
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
