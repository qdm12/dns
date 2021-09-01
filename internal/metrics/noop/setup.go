package noop

import (
	"context"

	cache "github.com/qdm12/dns/pkg/cache/metrics/noop"
	doh "github.com/qdm12/dns/pkg/doh/metrics/noop"
	dot "github.com/qdm12/dns/pkg/dot/metrics/noop"
)

type DummyRunner struct{}

func (d *DummyRunner) Run(ctx context.Context, done chan<- struct{}) {
	close(done)
}

func Setup() (dummy *DummyRunner,
	cacheMetrics *cache.Metrics,
	dotMetrics *dot.Metrics,
	dohMetrics *doh.Metrics) {
	cacheMetrics = cache.New()
	dotMetrics = dot.New()
	dohMetrics = doh.New()
	return new(DummyRunner), cacheMetrics, dotMetrics, dohMetrics
}
