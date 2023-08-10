package setup

import (
	"fmt"

	"github.com/qdm12/dns/v2/internal/config/settings"
	"github.com/qdm12/dns/v2/pkg/dot"
	noopmetrics "github.com/qdm12/dns/v2/pkg/dot/metrics/noop"
	prometheusmetrics "github.com/qdm12/dns/v2/pkg/dot/metrics/prometheus"
	"github.com/qdm12/dns/v2/pkg/metrics/prometheus"
)

func dotServer(userSettings settings.Settings,
	middlewares []Middleware, logger Logger, metrics DoTMetrics) (
	server *dot.Server, err error) {
	resolverSettings := dot.ResolverSettings{
		DoTProviders: userSettings.DoT.DoTProviders,
		DNSProviders: userSettings.DoT.DNSProviders,
		IPv6:         *userSettings.DoT.IPv6,
		Warner:       logger,
		Metrics:      metrics,
	}

	settings := dot.ServerSettings{
		Resolver:         resolverSettings,
		ListeningAddress: userSettings.ListeningAddress,
		Middlewares:      toDoTMiddlewares(middlewares),
		Logger:           logger,
	}

	return dot.NewServer(settings)
}

func dotMetrics(metricsType string, //nolint:ireturn
	commonPrometheus prometheus.Settings) (
	metrics DoTMetrics, err error) {
	switch metricsType {
	case noopString:
		return noopmetrics.New(), nil
	case prometheusString:
		prometheusSettings := prometheusmetrics.Settings{
			Prometheus: commonPrometheus,
		}
		return prometheusmetrics.New(prometheusSettings)
	default:
		panic(fmt.Sprintf("unknown metrics type: %s", metricsType))
	}
}

func toDoTMiddlewares(middlewares []Middleware) (dohMiddlewres []dot.Middleware) {
	dohMiddlewres = make([]dot.Middleware, len(middlewares))
	for i, middleware := range middlewares {
		dohMiddlewres[i] = dot.Middleware(middleware)
	}
	return dohMiddlewres
}
