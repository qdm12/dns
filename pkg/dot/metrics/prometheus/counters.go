package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/v2/internal/metrics/prometheus/helpers"
	prom "github.com/qdm12/dns/v2/pkg/metrics/prometheus"
)

type counters struct {
	dotDial *prometheus.CounterVec
}

func newCounters(settings prom.Settings) (c *counters, err error) {
	prefix := settings.Prefix
	c = &counters{
		dotDial: helpers.NewCounterVec(prefix, "dns_over_tls_dials",
			"DNS over TLS dials by provider, address and outcome", []string{"provider", "address", "outcome"}),
	}

	err = helpers.Register(settings.Registry, c.dotDial)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *counters) DoTDialInc(provider, address, outcome string) {
	c.dotDial.WithLabelValues(provider, address, outcome).Inc()
}
