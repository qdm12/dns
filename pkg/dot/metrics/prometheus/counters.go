package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/internal/metrics/prometheus/helpers"
	prom "github.com/qdm12/dns/pkg/metrics/prometheus"
)

type counters struct {
	dotDialProvider *prometheus.CounterVec
	dotDialAddress  *prometheus.CounterVec
	dnsDialProvider *prometheus.CounterVec
	dnsDialAddress  *prometheus.CounterVec
}

func newCounters(settings prom.Settings) (c *counters, err error) {
	c = &counters{
		dotDialProvider: helpers.NewCounterVec(settings.Prefix, "dot_dial_provider",
			"DNS over TLS dials by provider", []string{"provider", "outcome"}),
		dotDialAddress: helpers.NewCounterVec(settings.Prefix, "dot_dial_address",
			"DNS over TLS dials by address", []string{"address", "outcome"}),
		dnsDialProvider: helpers.NewCounterVec(settings.Prefix, "dns_dial_provider",
			"DNS dials by provider", []string{"provider", "outcome"}),
		dnsDialAddress: helpers.NewCounterVec(settings.Prefix, "dns_dial_address",
			"DNS dials by address", []string{"address", "outcome"}),
	}

	err = helpers.Register(settings.Registry, c.dotDialProvider, c.dotDialAddress,
		c.dnsDialProvider, c.dnsDialAddress)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *counters) DoTDialProviderInc(provider, outcome string) {
	c.dotDialProvider.WithLabelValues(provider, outcome).Inc()
}

func (c *counters) DoTDialAddressInc(address, outcome string) {
	c.dotDialAddress.WithLabelValues(address, outcome).Inc()
}

func (c *counters) DNSDialProviderInc(provider, outcome string) {
	c.dnsDialProvider.WithLabelValues(provider, outcome).Inc()
}

func (c *counters) DNSDialAddressInc(address, outcome string) {
	c.dnsDialAddress.WithLabelValues(address, outcome).Inc()
}
