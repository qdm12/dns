// Package metrics defines metric interfaces for the
// DoH server and resolver.
package metrics

import (
	"github.com/qdm12/dns/v2/pkg/doh/metrics/noop"
	"github.com/qdm12/dns/v2/pkg/doh/metrics/prometheus"
	dotmetrics "github.com/qdm12/dns/v2/pkg/dot/metrics"
)

var (
	_ DialMetrics = (*prometheus.Metrics)(nil)
	_ DialMetrics = (*noop.Metrics)(nil)
)

type DialMetrics interface {
	DoHDialMetrics
	dotmetrics.DialMetrics
}

type DoHDialMetrics interface {
	DoHDialInc(url string)
}
