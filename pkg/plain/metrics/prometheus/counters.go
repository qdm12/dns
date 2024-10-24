package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/v2/internal/metrics/prometheus/helpers"
	prom "github.com/qdm12/dns/v2/pkg/metrics/prometheus"
)

type counters struct {
	plainDial *prometheus.CounterVec
}

func newCounters(settings prom.Settings) (c *counters, err error) {
	prefix := settings.Prefix
	c = &counters{
		plainDial: helpers.NewCounterVec(prefix, "plain_dials",
			"Plain dials by address and outcome", []string{"address", "outcome"}),
	}

	err = helpers.Register(settings.Registry, c.plainDial)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *counters) PlainDialInc(address, outcome string) {
	c.plainDial.WithLabelValues(address, outcome).Inc()
}
