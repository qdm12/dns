package setup

import (
	"context"
	"fmt"

	"github.com/qdm12/dns/v2/internal/config/settings"
	"github.com/qdm12/dns/v2/pkg/metrics/prometheus"
	"github.com/qdm12/dns/v2/pkg/middlewares/log"
)

type Server interface {
	Run(ctx context.Context, stopped chan<- error)
}

func DNS(serverCtx context.Context, //nolint:ireturn
	userSettings settings.Settings,
	cache Cache, filter Filter,
	logger Logger, promRegistry PrometheusRegisterer) (
	server Server, err error) {
	middlewareLogger := makeMiddlewareLogger(logger, userSettings.Log)
	logMiddlewareSettings := log.Settings{
		// TODO formatter
		Logger: middlewareLogger,
	}

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

	switch userSettings.Upstream {
	case "dot":
		dotMetrics, err := dotMetrics(metricsType,
			commonPrometheus, middlewareMetrics)
		if err != nil {
			return nil, fmt.Errorf("DoT metrics: %w", err)
		}

		return dotServer(serverCtx, userSettings, logger,
			logMiddlewareSettings, dotMetrics, cache, filter)
	case "doh":
		dohMetrics, err := dohMetrics(metricsType,
			commonPrometheus, middlewareMetrics)
		if err != nil {
			return nil, fmt.Errorf("DoH metrics: %w", err)
		}

		return dohServer(serverCtx, userSettings, logger, dohMetrics,
			cache, filter)
	default:
		panic(fmt.Sprintf("unknown upstream: %s", userSettings.Upstream))
	}
}
