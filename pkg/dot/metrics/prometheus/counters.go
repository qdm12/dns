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
		dotDial: helpers.NewCounterVec(settings.Prefix, "dns_over_tls_dials",
			"DNS over TLS dials by provider, address and outcome", []string{"provider", "address", "outcome"}),
		dnsDial: helpers.NewCounterVec(settings.Prefix, "dns_plaintext_fallback_dials",
			"DNS dials by provider, address and outcome", []string{"address", "outcome"}),
	}

	err = helpers.Register(settings.Registry, c.dotDial, c.dnsDial)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *counters) DoTDialInc(provider, address, outcome string) {
	c.dotDial.WithLabelValues(provider, address, outcome).Inc()
}

func (c *counters) DNSDialInc(address, outcome string) {
	c.dnsDial.WithLabelValues(address, outcome).Inc()
}
