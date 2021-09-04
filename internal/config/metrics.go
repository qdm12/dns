package config

import (
	"fmt"

	filtermetrics "github.com/qdm12/dns/pkg/blacklist/metrics"
	cachemetrics "github.com/qdm12/dns/pkg/cache/metrics"
	dohmetrics "github.com/qdm12/dns/pkg/doh/metrics"
	dotmetrics "github.com/qdm12/dns/pkg/dot/metrics"
	"github.com/qdm12/golibs/params"
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

func getMetricsSettings(reader *reader) (settings Metrics,
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

func (m *Metrics) Lines(indent, subSection string) (lines []string) {
	lines = append(lines, subSection+"Type: "+m.Type)
	if m.Type == MetricProm {
		lines = append(lines, subSection+"Listening address: "+m.Prometheus.Address)
	}
	return lines
}
