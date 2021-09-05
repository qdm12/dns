package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/internal/metrics/prometheus/helpers"
	prom "github.com/qdm12/dns/pkg/metrics/prometheus"
)

type counters struct {
	dohDialURL *prometheus.CounterVec
}

func newCounters(settings prom.Settings) (c *counters, err error) {
	c = &counters{
		dohDialURL: helpers.NewCounterVec(settings.Prefix, "dns_over_https_dials",
			"DNS over HTTPS dials by URL", []string{"url"}),
	}

	err = helpers.Register(settings.Registry, c.dohDialURL)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *counters) DoHDialInc(url string) {
	c.dohDialURL.WithLabelValues(url).Inc()
}
