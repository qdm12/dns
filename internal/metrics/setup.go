// Package metrics sets up metrics and creates a metrics service.
package metrics

import (
	"fmt"

	dto "github.com/prometheus/client_model/go"
	"github.com/qdm12/dns/v2/internal/config/settings"
	"github.com/qdm12/dns/v2/internal/metrics/noop"
	"github.com/qdm12/dns/v2/internal/metrics/prometheus"
	"github.com/qdm12/log"
)

type ParentLogger interface {
	New(options ...log.Option) *log.Logger
}

type PrometheusGatherer interface {
	Gather() ([]*dto.MetricFamily, error)
}

type Service interface {
	String() string
	Start() (runError <-chan error, startErr error)
	Stop() (err error)
}

func New(settings settings.Metrics, //nolint:ireturn
	parentLogger ParentLogger, prometheusGatherer PrometheusGatherer) (
	service Service, err error) {
	switch settings.Type {
	case "noop":
		return noop.New()
	case "prometheus":
		logger := parentLogger.New(log.SetComponent("prometheus server"))
		return prometheus.New(settings.Prometheus, prometheusGatherer, logger)
	default:
		panic(fmt.Sprintf("unknown metrics type: %s", settings.Type))
	}
}
