package config

import (
	"fmt"

	cachemetrics "github.com/qdm12/dns/pkg/cache/metrics"
	dohmetrics "github.com/qdm12/dns/pkg/doh/metrics"
	dotmetrics "github.com/qdm12/dns/pkg/dot/metrics"
	filtermetrics "github.com/qdm12/dns/pkg/filter/metrics"
	"github.com/qdm12/golibs/params"
	"github.com/qdm12/gotree"
)

func (settings *Settings) PatchMetrics(
	cacheMetrics cachemetrics.Interface,
	filterMetrics filtermetrics.Interface,
	dotMetrics dotmetrics.Interface,
	dohMetrics dohmetrics.Interface) {
	settings.Cache.LRU.Metrics = cacheMetrics
	settings.Cache.Noop.Metrics = cacheMetrics
	settings.Filter.Metrics = filterMetrics
	settings.DoT.Metrics = dotMetrics
	settings.DoT.Resolver.Metrics = dotMetrics
	settings.DoH.Metrics = dohMetrics
	settings.DoH.Resolver.Metrics = dohMetrics
}

const (
	MetricNoop = "noop"
	MetricProm = "prometheus"
)

type Metrics struct {
	Type       string
	Prometheus Prometheus
}

func getMetricsSettings(reader *Reader) (settings Metrics,
	err error) {
	settings.Type, err = reader.env.Inside("METRICS_TYPE",
		[]string{MetricNoop, MetricProm}, params.Default("noop"))
	if err != nil {
		return settings, fmt.Errorf("environment variable METRICS_TYPE: %w", err)
	}

	settings.Prometheus, err = getPrometheusSettings(reader)
	if err != nil {
		return settings, err
	}

	return settings, nil
}

func (m *Metrics) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("Metrics settings:")
	if m.Type == MetricProm {
		node.AppendNode(m.Prometheus.ToLinesNode())
	}
	return node
}
