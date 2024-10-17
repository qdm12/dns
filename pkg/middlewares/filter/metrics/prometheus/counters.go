package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/v2/internal/metrics/prometheus/helpers"
	prom "github.com/qdm12/dns/v2/pkg/metrics/prometheus"
)

type counters struct {
	hostnamesFiltered *prometheus.CounterVec
	ipsFiltered       *prometheus.CounterVec
}

func newCounters(settings prom.Settings) (c *counters, err error) {
	prefix := settings.Prefix
	c = &counters{
		hostnamesFiltered: helpers.NewCounterVec(prefix,
			"hostnames_filtered",
			"DNS filtered out hostnames by question class and type",
			[]string{"class", "type"}),
		ipsFiltered: helpers.NewCounterVec(prefix,
			"ips_filtered",
			"IP addresses filtered out by response type",
			[]string{"type"}),
	}

	if err = helpers.Register(settings.Registry,
		c.hostnamesFiltered, c.ipsFiltered); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *counters) HostnamesFilteredInc(class, qType string) {
	c.hostnamesFiltered.WithLabelValues(class, qType).Inc()
}

func (c *counters) IPsFilteredInc(rrtype string) {
	c.ipsFiltered.WithLabelValues(rrtype).Inc()
}
