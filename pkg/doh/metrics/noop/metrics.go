// Package noop defines a No-op metric implementation for DoH.
package noop

import (
	dotmetrics "github.com/qdm12/dns/v2/pkg/dot/metrics"
	dotnoop "github.com/qdm12/dns/v2/pkg/dot/metrics/noop"
	middleware "github.com/qdm12/dns/v2/pkg/middlewares/metrics"
	middlewarenoop "github.com/qdm12/dns/v2/pkg/middlewares/metrics/noop"
)

type (
	dotDialMetrics      = dotmetrics.DialMetrics
	middlewareInterface = middleware.Interface
)

type Metrics struct {
	dotDialMetrics
	middlewareInterface
}

func New() (metrics *Metrics) {
	return &Metrics{
		dotDialMetrics:      dotnoop.New(),
		middlewareInterface: middlewarenoop.New(),
	}
}

func (m *Metrics) DoHDialInc(url string) {}
