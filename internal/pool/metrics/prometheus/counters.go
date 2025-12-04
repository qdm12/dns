package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/v2/internal/metrics/prometheus/helpers"
	prom "github.com/qdm12/dns/v2/pkg/metrics/prometheus"
)

type counters struct {
	newConns     *prometheus.CounterVec
	renewedConns *prometheus.CounterVec
	deadConns    *prometheus.CounterVec
	removedConns *prometheus.CounterVec
	getConns     *prometheus.CounterVec
	putConns     *prometheus.CounterVec
}

func newCounters(settings prom.Settings) (c *counters, err error) {
	prefix := settings.Prefix
	c = &counters{
		newConns: helpers.NewCounterVec(prefix, "pool_new_connections",
			"Pool new connections by address and outcome.", []string{"address", "outcome"}),
		renewedConns: helpers.NewCounterVec(prefix, "pool_renewed_connections",
			"Pool renewed connections by address, reason and outcome.", []string{"address", "reason", "outcome"}),
		deadConns: helpers.NewCounterVec(prefix, "pool_dead_connections",
			"Pool dead connections by address.", []string{"address"}),
		removedConns: helpers.NewCounterVec(prefix, "pool_removed_connections",
			"Pool removed connections by address.", []string{"address"}),
		getConns: helpers.NewCounterVec(prefix, "pool_get_connection",
			"Pool get connection calls by address and outcome.", []string{"address", "outcome"}),
		putConns: helpers.NewCounterVec(prefix, "pool_put_connection",
			"Pool put connection calls by address and connection state, live or dead.",
			[]string{"address", "state"}),
	}

	err = helpers.Register(settings.Registry, c.newConns, c.renewedConns,
		c.deadConns, c.removedConns, c.getConns, c.putConns)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *counters) NewConnsInc(address, outcome string) {
	c.newConns.WithLabelValues(address, outcome).Inc()
}

func (c *counters) RenewedConnsInc(address, reason, outcome string) {
	c.renewedConns.WithLabelValues(address, reason, outcome).Inc()
}

func (c *counters) DeadConnsInc(address string) {
	c.deadConns.WithLabelValues(address).Inc()
}

func (c *counters) RemovedConnsAdd(address string, removed uint) {
	c.removedConns.WithLabelValues(address).Add(float64(removed))
}

func (c *counters) GetConnsInc(address, outcome string) {
	c.getConns.WithLabelValues(address, outcome).Inc()
}

func (c *counters) PutConnsInc(address, state string) {
	c.putConns.WithLabelValues(address, state).Inc()
}
