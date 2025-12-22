package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/v2/internal/metrics/prometheus/helpers"
	prom "github.com/qdm12/dns/v2/pkg/metrics/prometheus"
)

type counters struct {
	liveConns    *prometheus.CounterVec
	deadConns    *prometheus.CounterVec
	removedConns *prometheus.CounterVec
	getConns     *prometheus.CounterVec
	putConns     *prometheus.CounterVec
	newConns     *prometheus.CounterVec
	renewedConns *prometheus.CounterVec
}

func newCounters(settings prom.Settings) (c *counters, err error) {
	prefix := settings.Prefix
	c = &counters{
		liveConns: helpers.NewCounterVec(prefix, "pool_live_connections",
			"Pool live connections by address", []string{"address"}),
		deadConns: helpers.NewCounterVec(prefix, "pool_dead_connections",
			"Pool dead connections by address", []string{"address"}),
		removedConns: helpers.NewCounterVec(prefix, "pool_removed_connections",
			"Connections removed from the pool by address", []string{"address"}),
		getConns: helpers.NewCounterVec(prefix, "pool_get_connection",
			"Pool get connection operations by address and outcome", []string{"address", "outcome"}),
		putConns: helpers.NewCounterVec(prefix, "pool_put_connection",
			"Pool put connection operations by address and connection state, live or dead",
			[]string{"address", "state"}),
		newConns: helpers.NewCounterVec(prefix, "pool_new_connection",
			"Pool new connection operations by address and outcome", []string{"address", "outcome"}),
		renewedConns: helpers.NewCounterVec(prefix, "pool_renew_connection",
			"Pool renew connection operations by address, reason and outcome", []string{"address", "reason", "outcome"}),
	}

	err = helpers.Register(settings.Registry, c.liveConns, c.deadConns,
		c.removedConns, c.getConns, c.putConns, c.newConns, c.renewedConns)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (h *counters) LiveConnInc(address string) {
	h.liveConns.WithLabelValues(address).Inc()
}

func (h *counters) DeadConnInc(address string) {
	h.deadConns.WithLabelValues(address).Inc()
}

func (h *counters) RemovedConnsAdd(address string, removed uint) {
	h.removedConns.WithLabelValues(address).Add(float64(removed))
}

func (h *counters) GetConnInc(address, outcome string) {
	h.getConns.WithLabelValues(address, outcome).Inc()
}

func (h *counters) PutConnInc(address, state string) {
	h.putConns.WithLabelValues(address, state).Inc()
}

func (h *counters) NewConnsInc(address, outcome string) {
	h.newConns.WithLabelValues(address, outcome).Inc()
}

func (h *counters) RenewConnInc(address, reason, outcome string) {
	h.renewedConns.WithLabelValues(address, reason, outcome).Inc()
}
