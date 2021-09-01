// Package metrics sets up metrics and patch the given settings
package metrics

import (
	"context"

	"github.com/qdm12/dns/internal/config"
	"github.com/qdm12/dns/internal/metrics/noop"
	"github.com/qdm12/dns/internal/metrics/prometheus"
	"github.com/qdm12/golibs/logging"
)

type Runner interface {
	Run(ctx context.Context, done chan<- struct{})
}

func Setup(settings *config.Settings, parentLogger logging.ParentLogger) (
	runner Runner, err error,
) {
	if settings.Metrics.Type == config.MetricProm {
		return setupPrometheus(settings, parentLogger)
	}
	runner = setupNoop(settings)
	return runner, nil
}

func setupPrometheus(settings *config.Settings, parentLogger logging.ParentLogger) (
	runner *prometheus.Server, err error) {
	loggerSettings := logging.Settings{Prefix: "prometheus server"}
	logger := parentLogger.NewChild(loggerSettings)
	promServer, cacheMetrics, dotMetrics, dohMetrics, err :=
		prometheus.Setup(settings.Metrics.Prometheus, logger)
	if err != nil {
		return nil, err
	}

	settings.PatchMetrics(cacheMetrics, dotMetrics, dohMetrics)

	return promServer, nil
}

func setupNoop(settings *config.Settings) (runner *noop.DummyRunner) {
	runner, cacheMetrics, dotMetrics, dohMetrics := noop.Setup()
	settings.PatchMetrics(cacheMetrics, dotMetrics, dohMetrics)
	return runner
}
