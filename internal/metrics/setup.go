// Package metrics sets up metrics and patch the given settings
package metrics

import (
	"context"
	"fmt"

	dto "github.com/prometheus/client_model/go"
	"github.com/qdm12/dns/v2/internal/config/settings"
	"github.com/qdm12/dns/v2/internal/metrics/noop"
	"github.com/qdm12/dns/v2/internal/metrics/prometheus"
	"github.com/qdm12/golibs/logging"
)

type Runner interface {
	Run(ctx context.Context, done chan<- struct{})
}

type PrometheusGatherer interface {
	Gather() ([]*dto.MetricFamily, error)
}

func Setup(settings settings.Metrics, //nolint:ireturn
	parentLogger logging.ParentLogger,
	prometheusGatherer PrometheusGatherer) (runner Runner) {
	switch settings.Type {
	case "noop":
		return noop.Setup()
	case "prometheus":
		loggerSettings := logging.Settings{Prefix: "prometheus server: "}
		logger := parentLogger.NewChild(loggerSettings)
		return prometheus.Setup(settings.Prometheus, prometheusGatherer, logger)
	default:
		panic(fmt.Sprintf("unknown metrics type: %s", settings.Type))
	}
}
