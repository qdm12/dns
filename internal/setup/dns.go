package setup

import (
	"fmt"

	"github.com/qdm12/dns/v2/internal/config"
	"github.com/qdm12/dns/v2/pkg/metrics/prometheus"
	cachemiddleware "github.com/qdm12/dns/v2/pkg/middlewares/cache"
	filtermiddleware "github.com/qdm12/dns/v2/pkg/middlewares/filter"
	"github.com/qdm12/dns/v2/pkg/middlewares/localdns"
	"github.com/qdm12/log"
)

type Service interface {
	String() string
	Start() (runError <-chan error, startErr error)
	Stop() (err error)
}

func DNS(userSettings config.Settings, ipv6Support bool, //nolint:ireturn
	cache Cache, filter Filter, loggerConstructor LoggerConstructor,
	promRegistry PrometheusRegistry) (server Service, err error) {
	commonPrometheus := prometheus.Settings{
		Prefix:   *userSettings.Metrics.Prometheus.Subsystem,
		Registry: promRegistry,
	}

	middlewares, err := setupMiddlewares(userSettings, cache,
		filter, loggerConstructor, commonPrometheus)
	if err != nil {
		return nil, fmt.Errorf("setting up middlewares: %w", err)
	}

	switch userSettings.Upstream {
	case "dot":
		logger := loggerConstructor.New(log.SetComponent("DNS over TLS"))
		dotMetrics, err := dotMetrics(userSettings.Metrics.Type, commonPrometheus)
		if err != nil {
			return nil, fmt.Errorf("DoT metrics: %w", err)
		}

		return dotServer(userSettings, ipv6Support, middlewares,
			logger, dotMetrics)
	case "doh":
		logger := loggerConstructor.New(log.SetComponent("DNS over HTTPS"))
		dohMetrics, err := dohMetrics(userSettings.Metrics.Type, commonPrometheus)
		if err != nil {
			return nil, fmt.Errorf("DoH metrics: %w", err)
		}

		return dohServer(userSettings, ipv6Support, middlewares,
			logger, dohMetrics)
	default:
		panic(fmt.Sprintf("unknown upstream: %s", userSettings.Upstream))
	}
}

func setupMiddlewares(userSettings config.Settings, cache Cache,
	filter Filter, loggerConstructor log.ChildConstructor, commonPrometheus prometheus.Settings) (
	middlewares []Middleware, err error) {
	cacheMiddleware, err := cachemiddleware.New(cachemiddleware.Settings{Cache: cache})
	if err != nil {
		return nil, fmt.Errorf("creating cache middleware: %w", err)
	}
	middlewares = append(middlewares, cacheMiddleware)

	if *userSettings.LocalDNS.Enabled {
		localDNSMiddleware, err := localdns.New(localdns.Settings{
			Resolvers: userSettings.LocalDNS.Resolvers, // possibly auto-detected
			Logger:    loggerConstructor.New(log.SetComponent("local redirector")),
		})
		if err != nil {
			return nil, fmt.Errorf("creating local DNS middleware: %w", err)
		}
		// Place after cache middleware, since we want to avoid caching for local
		// hostnames that may change regularly.
		middlewares = append(middlewares, localDNSMiddleware)
	}

	filterMiddleware, err := filtermiddleware.New(filtermiddleware.Settings{Filter: filter})
	if err != nil {
		return nil, fmt.Errorf("creating filter middleware: %w", err)
	}
	// Note the filter middleware must be wrapping the cache middleware
	// to catch filtered responses found from the cache.
	middlewares = append(middlewares, filterMiddleware)

	metricsMiddleware, err := middlewareMetrics(userSettings.Metrics.Type,
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

	return middlewares, nil
}
