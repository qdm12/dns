package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/internal/metrics/prometheus/helpers"
	prom "github.com/qdm12/dns/pkg/prometheus"
)

type counters struct {
	dohDialURL *prometheus.CounterVec
}

func newCounters(settings prom.Settings) (c *counters, err error) {
	c = &counters{
		dohDialURL: helpers.NewCounterVec(settings.Prefix, "doh_dial_url",
			"DNS over HTTPS dials by URL", []string{"url"}),
	}

	err = helpers.Register(settings.Registry, c.dohDialURL)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *counters) DoHDialURLInc(url string) {
	c.dohDialURL.WithLabelValues(url).Inc()
}
