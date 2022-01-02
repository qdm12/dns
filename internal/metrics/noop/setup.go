// Package noop initializes all No-op metrics objects.
package noop

import (
	"context"

	cache "github.com/qdm12/dns/v2/pkg/cache/metrics/noop"
	doh "github.com/qdm12/dns/v2/pkg/doh/metrics/noop"
	dot "github.com/qdm12/dns/v2/pkg/dot/metrics/noop"
	filter "github.com/qdm12/dns/v2/pkg/filter/metrics/noop"
)

type DummyRunner struct{}

func (d *DummyRunner) Run(ctx context.Context, done chan<- struct{}) {
	close(done)
}

func Setup() (dummy *DummyRunner,
	cacheMetrics *cache.Metrics,
	filterMetrics *filter.Metrics,
	dotMetrics *dot.Metrics,
	dohMetrics *doh.Metrics) {
	cacheMetrics = cache.New()
	filterMetrics = filter.New()
	dotMetrics = dot.New()
	dohMetrics = doh.New()
	return new(DummyRunner), cacheMetrics, filterMetrics, dotMetrics, dohMetrics
}
