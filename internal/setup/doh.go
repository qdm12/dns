package setup

import (
	"fmt"

	"github.com/qdm12/dns/v2/internal/config"
	"github.com/qdm12/dns/v2/pkg/doh"
	noopmetrics "github.com/qdm12/dns/v2/pkg/doh/metrics/noop"
	prometheusmetrics "github.com/qdm12/dns/v2/pkg/doh/metrics/prometheus"
	"github.com/qdm12/dns/v2/pkg/metrics/prometheus"
	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/qdm12/gosettings"
)

func dohServer(userSettings config.Settings,
	middlewares []Middleware, logger Logger, metrics DoHMetrics) (
	server *doh.Server, err error) {
	providers := provider.NewProviders()

	DoHProviders, err := stringsToProviders(providers, userSettings.DoH.DoHProviders)
	if err != nil {
		return nil, fmt.Errorf("DNS over HTTPS providers: %w", err)
	}

	selfDoTProviders, err := stringsToProviders(providers, userSettings.DoT.DoTProviders)
	if err != nil {
		return nil, fmt.Errorf("DNS over TLS providers: %w", err)
	}

	selfDNSProviders, err := stringsToProviders(providers, userSettings.DoT.DNSProviders)
	if err != nil {
		return nil, fmt.Errorf("plaintext DNS providers: %w", err)
	}

	resolverSettings := doh.ResolverSettings{
		DoHProviders: DoHProviders,
		SelfDNS: doh.SelfDNS{
			DoTProviders: selfDoTProviders,
			DNSProviders: selfDNSProviders,
			Timeout:      userSettings.DoH.Self.Timeout,
			IPv6:         *userSettings.DoH.Self.IPv6,
		},
		Warner:  logger,
		Metrics: metrics,
	}

	settings := doh.ServerSettings{
		Resolver:         resolverSettings,
		ListeningAddress: gosettings.CopyPointer(userSettings.ListeningAddress),
		Middlewares:      toDoHMiddlewares(middlewares),
		Logger:           logger,
	}

	return doh.NewServer(settings)
}

func dohMetrics(metricsType string, //nolint:ireturn
	commonPrometheus prometheus.Settings) (
	metrics DoHMetrics, err error) {
	switch metricsType {
	case noopString:
		return noopmetrics.New(), nil
	case prometheusString:
		dotDialMetrics, err := dotMetrics(metricsType, commonPrometheus)
		if err != nil {
			return nil, fmt.Errorf("DoT metrics: %w", err)
		}

		prometheusSettings := prometheusmetrics.Settings{
			Prometheus:     commonPrometheus,
			DoTDialMetrics: dotDialMetrics,
		}
		return prometheusmetrics.New(prometheusSettings)
	default:
		panic(fmt.Sprintf("unknown metrics type: %s", metricsType))
	}
}

func toDoHMiddlewares(middlewares []Middleware) (dohMiddlewres []doh.Middleware) {
	dohMiddlewres = make([]doh.Middleware, len(middlewares))
	for i, middleware := range middlewares {
		dohMiddlewres[i] = doh.Middleware(middleware)
	}
	return dohMiddlewres
}

func stringsToProviders(providers *provider.Providers, providerNames []string) (
	providersSlice []provider.Provider, err error) {
	providersSlice = make([]provider.Provider, len(providerNames))
	for i, providerName := range providerNames {
		providersSlice[i], err = providers.Get(providerName)
		if err != nil {
			return nil, fmt.Errorf("provider %s: %w", providerName, err)
		}
	}
	return providersSlice, nil
}
