package setup

import (
	"fmt"

	"github.com/qdm12/dns/v2/internal/config/settings"
	"github.com/qdm12/dns/v2/pkg/metrics/prometheus"
)

type Service interface {
	String() string
	Start() (runError <-chan error, startErr error)
	Stop() (err error)
}

func DNS(userSettings settings.Settings, //nolint:ireturn
	cache Cache, filter Filter, logger Logger, promRegistry PrometheusRegisterer) (
	server Service, err error) {
	var middlewares []Middleware

	commonPrometheus := prometheus.Settings{
		Prefix:   *userSettings.Metrics.Prometheus.Subsystem,
		Registry: promRegistry,
	}

	metricsType := userSettings.Metrics.Type

	middlewareMetrics, err := middlewareMetrics(metricsType,
		commonPrometheus)
	if err != nil {
		return nil, fmt.Errorf("middleware metrics: %w", err)
	}
	middlewares = append(middlewares, middlewareMetrics)

	logMiddleware, err := logMiddleware(userSettings.MiddlewareLog)
	if err != nil {
		return nil, fmt.Errorf("log middleware: %w", err)
	}
	middlewares = append(middlewares, logMiddleware)

	switch userSettings.Upstream {
	case "dot":
		dotMetrics, err := dotMetrics(metricsType, commonPrometheus)
		if err != nil {
			return nil, fmt.Errorf("DoT metrics: %w", err)
		}

		return dotServer(userSettings, middlewares,
			logger, dotMetrics, cache, filter)
	case "doh":
		dohMetrics, err := dohMetrics(metricsType, commonPrometheus)
		if err != nil {
			return nil, fmt.Errorf("DoH metrics: %w", err)
		}

		return dohServer(userSettings, middlewares,
			logger, dohMetrics, cache, filter)
	default:
		panic(fmt.Sprintf("unknown upstream: %s", userSettings.Upstream))
	}
}
