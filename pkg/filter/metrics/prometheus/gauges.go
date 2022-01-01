package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/internal/metrics/prometheus/helpers"
	prom "github.com/qdm12/dns/pkg/metrics/prometheus"
)

type gauges struct {
	blockedHostnames  prometheus.Gauge
	blockedIPs        prometheus.Gauge
	blockedIPPrefixes prometheus.Gauge
}

func newGauges(settings prom.Settings) (g *gauges, err error) {
	prefix := *settings.Prefix
	g = &gauges{
		blockedHostnames: helpers.NewGauge(prefix,
			"blocked_hostnames", "Total number of hostnames to be blocked by the DNS server filter"),
		blockedIPs: helpers.NewGauge(prefix,
			"blocked_ips", "Total number of IP addresses to be blocked by the DNS server filter"),
		blockedIPPrefixes: helpers.NewGauge(prefix,
			"blocked_ip_prefixes", "Total number of IP address prefixes to be blocked by the DNS server filter"),
	}

	err = helpers.Register(settings.Registry,
		g.blockedHostnames, g.blockedIPs, g.blockedIPPrefixes)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (g *gauges) SetBlockedHostnames(n int)  { g.blockedHostnames.Set(float64(n)) }
func (g *gauges) SetBlockedIPs(n int)        { g.blockedIPs.Set(float64(n)) }
func (g *gauges) SetBlockedIPPrefixes(n int) { g.blockedIPPrefixes.Set(float64(n)) }
