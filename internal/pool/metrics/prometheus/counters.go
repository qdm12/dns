package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/v2/internal/metrics/prometheus/helpers"
	prom "github.com/qdm12/dns/v2/pkg/metrics/prometheus"
)

type counters struct {
	renewRequests *prometheus.CounterVec
	renewals      *prometheus.CounterVec
}

func newCounters(settings prom.Settings) (c *counters, err error) {
	prefix := settings.Prefix
	c = &counters{
		renewRequests: helpers.NewCounterVec(prefix, "pool_renew_requests",
			"Pool Connection renew requests by address. "+
				"This is generally triggered following a connection error.", []string{"address"}),
		renewals: helpers.NewCounterVec(prefix, "pool_connection_renewals",
			"Pool connection renewals by address. "+
				"This can be caused either by a connection error or "+
				"internally when a connection is assumed expired.", []string{"address"}),
	}

	err = helpers.Register(settings.Registry, c.renewRequests, c.renewals)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *counters) RenewRequestsInc(address string) {
	c.renewRequests.WithLabelValues(address).Inc()
}

func (c *counters) RenewalsInc(address string) {
	c.renewals.WithLabelValues(address).Inc()
}
