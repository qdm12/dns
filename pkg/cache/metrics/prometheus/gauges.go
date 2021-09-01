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
	g = &gauges{
		maxEntries: helpers.NewGauge(settings.Prefix, "cache_max_entries", "DNS cache maximum number of entries"),
	}

	countersToRegister := []prometheus.Gauge{g.maxEntries}
	for _, gauge := range countersToRegister {
		if err = settings.Registry.Register(gauge); err != nil {
			return g, err
		}
	}

	return g, nil
}

func (g *gauges) CacheMaxEntriesSet(maxEntries int) {
	g.maxEntries.Set(float64(maxEntries))
}
