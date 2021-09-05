// Package metrics defines a metrics interface for the blacklister.
package metrics

import (
	"github.com/qdm12/dns/pkg/blacklist/metrics/noop"
	"github.com/qdm12/dns/pkg/blacklist/metrics/prometheus"
)

var (
	_ Interface = (*prometheus.Metrics)(nil)
	_ Interface = (*noop.Metrics)(nil)
)

type Interface interface {
	SetBlockedHostnames(n int)
	SetBlockedIPs(n int)
	SetBlockedIPPrefixes(n int)
	HostnamesFilteredInc(qClass, qType string)
	IPsFilteredInc(rrtype string)
}
