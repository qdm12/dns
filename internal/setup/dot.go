package setup

import (
	"context"
	"fmt"

	"github.com/qdm12/dns/v2/internal/config/settings"
	"github.com/qdm12/dns/v2/pkg/dot"
	noopmetrics "github.com/qdm12/dns/v2/pkg/dot/metrics/noop"
	prometheusmetrics "github.com/qdm12/dns/v2/pkg/dot/metrics/prometheus"
	"github.com/qdm12/dns/v2/pkg/metrics/prometheus"
	"github.com/qdm12/dns/v2/pkg/middlewares/log"
)

func dotServer(serverCtx context.Context,
	userSettings settings.Settings,
	logger Logger, logMiddlewareSettings log.Settings,
	metrics DoTMetrics,
	cache Cache, filter Filter) (server *dot.Server, err error) {
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
		LogMiddleware:    logMiddlewareSettings,
		Cache:            cache,
		Filter:           filter,
		Logger:           logger,
		Metrics:          metrics,
	}

	return dot.NewServer(serverCtx, settings)
}

func dotMetrics(metricsType string, //nolint:ireturn
	commonPrometheus prometheus.Settings,
	middleware MiddlewareMetrics) (
	metrics DoTMetrics, err error) {
	switch metricsType {
	case noopString:
		return noopmetrics.New(), nil
	case prometheusString:
		prometheusSettings := prometheusmetrics.Settings{
			Prometheus:        commonPrometheus,
			MiddlewareMetrics: middleware,
		}
		return prometheusmetrics.New(prometheusSettings)
	default:
		panic(fmt.Sprintf("unknown metrics type: %s", metricsType))
	}
}
