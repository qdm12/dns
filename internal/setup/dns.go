package setup

import (
	"fmt"

	"github.com/qdm12/dns/v2/internal/config/settings"
	"github.com/qdm12/dns/v2/pkg/metrics/prometheus"
	"github.com/qdm12/dns/v2/pkg/middlewares/log"
)

type Service interface {
	String() string
	Start() (runError <-chan error, startErr error)
	Stop() (err error)
}

func DNS(userSettings settings.Settings, //nolint:ireturn
	cache Cache, filter Filter, logger Logger, promRegistry PrometheusRegisterer,
	logMiddlewareSettings log.Settings) (
	server Service, err error) {
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

		return dotServer(userSettings, logger,
			logMiddlewareSettings, dotMetrics, cache, filter)
	case "doh":
		dohMetrics, err := dohMetrics(metricsType,
			commonPrometheus, middlewareMetrics)
		if err != nil {
			return nil, fmt.Errorf("DoH metrics: %w", err)
		}

		return dohServer(userSettings, logger,
			logMiddlewareSettings, dohMetrics, cache, filter)
	default:
		panic(fmt.Sprintf("unknown upstream: %s", userSettings.Upstream))
	}
}
