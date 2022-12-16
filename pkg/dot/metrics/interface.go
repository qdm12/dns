// Package metrics defines metric interfaces for the
// DoT server and resolver.
package metrics

import (
	"github.com/qdm12/dns/v2/pkg/dot/metrics/noop"
	"github.com/qdm12/dns/v2/pkg/dot/metrics/prometheus"
)

var (
	_ DialMetrics = (*prometheus.Metrics)(nil)
	_ DialMetrics = (*noop.Metrics)(nil)
)

type DialMetrics interface {
	DoTDialInc(provider, address, outcome string)
	DialDNSMetrics
}

type DialDNSMetrics interface {
	DNSDialInc(address, outcome string)
}
