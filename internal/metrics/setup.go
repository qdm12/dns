// Package metrics sets up metrics and patch the given settings
package metrics

import (
	"context"
	"fmt"

	dto "github.com/prometheus/client_model/go"
	"github.com/qdm12/dns/v2/internal/config/settings"
	"github.com/qdm12/dns/v2/internal/metrics/noop"
	"github.com/qdm12/dns/v2/internal/metrics/prometheus"
	"github.com/qdm12/log"
)

type Runner interface {
	Run(ctx context.Context, done chan<- struct{})
}

type Logger interface {
	New(options ...log.Option) *log.Logger
	Debug(s string)
	Info(s string)
	Warn(s string)
	Error(s string)
}

type PrometheusGatherer interface {
	Gather() ([]*dto.MetricFamily, error)
}

func Setup(settings settings.Metrics, //nolint:ireturn
	parentLogger Logger,
	prometheusGatherer PrometheusGatherer) (runner Runner) {
	switch settings.Type {
	case "noop":
		return noop.Setup()
	case "prometheus":
		logger := parentLogger.New(log.SetComponent("prometheus server: "))
		return prometheus.Setup(settings.Prometheus, prometheusGatherer, logger)
	default:
		panic(fmt.Sprintf("unknown metrics type: %s", settings.Type))
	}
}
