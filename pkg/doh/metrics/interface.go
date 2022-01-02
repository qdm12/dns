// Package metrics defines metric interfaces for the
// DoH server and resolver.
package metrics

import (
	"github.com/qdm12/dns/v2/pkg/doh/metrics/noop"
	"github.com/qdm12/dns/v2/pkg/doh/metrics/prometheus"
	dotmetrics "github.com/qdm12/dns/v2/pkg/dot/metrics"
	middleware "github.com/qdm12/dns/v2/pkg/middlewares/metrics"
)

var (
	_ Interface = (*prometheus.Metrics)(nil)
	_ Interface = (*noop.Metrics)(nil)
)

type Interface interface {
	DialMetrics
	middleware.Interface
}

type DialMetrics interface {
	DoHDialMetrics
	dotmetrics.DialMetrics
}

type DoHDialMetrics interface {
	DoHDialInc(url string)
}
