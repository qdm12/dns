package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/internal/metrics/prometheus/helpers"
	prom "github.com/qdm12/dns/pkg/metrics/prometheus"
)

type gauges struct {
	maxEntries prometheus.Gauge
}

func newGauges(settings prom.Settings) (g *gauges, err error) {
	prefix := *settings.Prefix
	g = &gauges{
		maxEntries: helpers.NewGauge(prefix, "cache_max_entries", "DNS cache maximum number of entries"),
	}

	err = helpers.Register(settings.Registry, g.maxEntries)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (g *gauges) CacheMaxEntriesSet(maxEntries int) {
	g.maxEntries.Set(float64(maxEntries))
}
