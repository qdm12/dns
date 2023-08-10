package setup

import (
	"fmt"

	"github.com/qdm12/dns/v2/internal/config/settings"
	cachemiddleware "github.com/qdm12/dns/v2/pkg/cache/middleware"
	filtermiddleware "github.com/qdm12/dns/v2/pkg/filter/middleware"
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

	middlewares = append(middlewares, cachemiddleware.New(cache))
	// Note the filter middleware must be wrapping the cache middleware
	// to catch filtered responses found from the cache.
	middlewares = append(middlewares, filtermiddleware.New(filter, cache))

	commonPrometheus := prometheus.Settings{
		Prefix:   *userSettings.Metrics.Prometheus.Subsystem,
		Registry: promRegistry,
	}

	metricsType := userSettings.Metrics.Type

	metricsMiddleware, err := middlewareMetrics(metricsType,
		commonPrometheus)
	if err != nil {
		return nil, fmt.Errorf("middleware metrics: %w", err)
	}
	middlewares = append(middlewares, metricsMiddleware)

	// Log middleware should be one of the top most middlewares
	// so it actually calls `.WriteMsg` on an actual dns.ResponseWriter
	// writing to the network. Having it as the last element of the
	// middlewares slice achieves this.
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
			logger, dotMetrics)
	case "doh":
		dohMetrics, err := dohMetrics(metricsType, commonPrometheus)
		if err != nil {
			return nil, fmt.Errorf("DoH metrics: %w", err)
		}

		return dohServer(userSettings, middlewares,
			logger, dohMetrics)
	default:
		panic(fmt.Sprintf("unknown upstream: %s", userSettings.Upstream))
	}
}
