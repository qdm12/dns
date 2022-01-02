package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/v2/internal/metrics/prometheus/helpers"
	prom "github.com/qdm12/dns/v2/pkg/metrics/prometheus"
)

type labels struct {
	cacheLabels *prometheus.GaugeVec
}

func newLabels(settings prom.Settings) (l *labels, err error) {
	prefix := *settings.Prefix
	l = &labels{
		cacheLabels: helpers.NewGaugeVec(
			prefix, "cache_labels", "DNS cache labels", []string{"type"}),
	}

	err = helpers.Register(settings.Registry, l.cacheLabels)
	if err != nil {
		return nil, err
	}

	return l, nil
}

func (l *labels) SetCacheType(cacheType string) {
	l.cacheLabels.WithLabelValues(cacheType).Set(0)
}
