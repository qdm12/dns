package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/internal/metrics/prometheus/helpers"
	prom "github.com/qdm12/dns/pkg/metrics/prometheus"
)

type counters struct {
	dotDial *prometheus.CounterVec
	dnsDial *prometheus.CounterVec
}

func newCounters(settings prom.Settings) (c *counters, err error) {
	c = &counters{
		dotDial: helpers.NewCounterVec(settings.Prefix, "dot_dial",
			"DNS over TLS dials by provider, address and outcome", []string{"provider", "address", "outcome"}),
		dnsDial: helpers.NewCounterVec(settings.Prefix, "dns_dial_provider",
			"DNS dials by provider, address and outcome", []string{"provider", "address", "outcome"}),
	}

	err = helpers.Register(settings.Registry, c.dotDial, c.dnsDial)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *counters) DoTDialInc(provider, address, outcome string) {
	c.dotDial.WithLabelValues(provider, outcome).Inc()
}

func (c *counters) DNSDialInc(provider, address, outcome string) {
	c.dnsDial.WithLabelValues(provider, outcome).Inc()
}
