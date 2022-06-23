package setup

import (
	"context"
	"fmt"

	"github.com/qdm12/dns/v2/internal/config/settings"
	"github.com/qdm12/dns/v2/pkg/doh"
	noopmetrics "github.com/qdm12/dns/v2/pkg/doh/metrics/noop"
	prometheusmetrics "github.com/qdm12/dns/v2/pkg/doh/metrics/prometheus"
	"github.com/qdm12/dns/v2/pkg/metrics/prometheus"
	"github.com/qdm12/dns/v2/pkg/middlewares/log"
)

func dohServer(serverCtx context.Context,
	userSettings settings.Settings,
	logger Logger, metrics DoHMetrics,
	cache Cache, filter Filter) (server *doh.Server, err error) {
	resolverSettings := doh.ResolverSettings{
		DoHProviders: userSettings.DoH.DoHProviders,
		SelfDNS: doh.SelfDNS{
			DoTProviders: userSettings.DoH.Self.DoTProviders,
			DNSProviders: userSettings.DoH.Self.DNSProviders,
			Timeout:      userSettings.DoH.Self.Timeout,
			IPv6:         *userSettings.DoH.Self.IPv6,
		},
		Warner:  logger,
		Metrics: metrics,
	}

	logMiddlewareSettings := log.Settings{}

	settings := doh.ServerSettings{
		Resolver:         resolverSettings,
		ListeningAddress: userSettings.ListeningAddress,
		LogMiddleware:    logMiddlewareSettings,
		Cache:            cache,
		Filter:           filter,
		Logger:           logger,
		Metrics:          metrics,
	}

	return doh.NewServer(serverCtx, settings)
}

func dohMetrics(metricsType string, //nolint:ireturn
	commonPrometheus prometheus.Settings,
	middleware MiddlewareMetrics) (
	metrics DoHMetrics, err error) {
	switch metricsType {
	case noopString:
		return noopmetrics.New(), nil
	case prometheusString:
		dotDialMetrics, err := dotMetrics(metricsType,
			commonPrometheus, middleware)
		if err != nil {
			return nil, fmt.Errorf("DoT metrics: %w", err)
		}

		prometheusSettings := prometheusmetrics.Settings{
			Prometheus:        commonPrometheus,
			DoTDialMetrics:    dotDialMetrics,
			MiddlewareMetrics: middleware,
		}
		return prometheusmetrics.New(prometheusSettings)
	default:
		panic(fmt.Sprintf("unknown metrics type: %s", metricsType))
	}
}
