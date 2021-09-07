// Package metrics defines metric interfaces for the
// DoT server and resolver.
package metrics

import (
	"github.com/qdm12/dns/pkg/dot/metrics/noop"
	"github.com/qdm12/dns/pkg/dot/metrics/prometheus"
	middleware "github.com/qdm12/dns/pkg/middlewares/metrics"
)

var (
	_ Interface = (*prometheus.Metrics)(nil)
	_ Interface = (*noop.Metrics)(nil)
)

//go:generate mockgen -destination=mock_$GOPACKAGE/$GOFILE . Interface

type Interface interface {
	DialMetrics
	middleware.Interface
}

type DialMetrics interface {
	DoTDialInc(provider, address, outcome string)
	DialDNSMetrics
}

type DialDNSMetrics interface {
	DNSDialInc(provider, address, outcome string)
}
