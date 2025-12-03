package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/v2/internal/metrics/prometheus/helpers"
	prom "github.com/qdm12/dns/v2/pkg/metrics/prometheus"
)

type gauges struct {
	conns      *prometheus.GaugeVec
	liveConns  *prometheus.GaugeVec
	inUseConns *prometheus.GaugeVec
}

func newGauges(settings prom.Settings) (g *gauges, err error) {
	prefix := settings.Prefix
	g = &gauges{
		conns: helpers.NewGaugeVec(prefix, "pool_connections",
			"Connections in the pool, both live and dead, by address", []string{"address"}),
		liveConns: helpers.NewGaugeVec(prefix, "pool_live_connections",
			"Live connections in the pool by address", []string{"address"}),
		inUseConns: helpers.NewGaugeVec(prefix, "pool_in_use_connections",
			"In use connections by address", []string{"address"}),
	}

	err = helpers.Register(settings.Registry, g.conns, g.liveConns, g.inUseConns)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (c *gauges) ConnsAdd(address string, n int) {
	c.conns.WithLabelValues(address).Add(float64(n))
}

func (c *gauges) LiveConnsAdd(address string, n int) {
	c.liveConns.WithLabelValues(address).Add(float64(n))
}

func (c *gauges) InUseConnsAdd(address string, n int) {
	c.inUseConns.WithLabelValues(address).Add(float64(n))
}
