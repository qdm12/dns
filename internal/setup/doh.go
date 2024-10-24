package setup

import (
	"fmt"

	"github.com/qdm12/dns/v2/internal/config"
	"github.com/qdm12/dns/v2/pkg/doh"
	noopmetrics "github.com/qdm12/dns/v2/pkg/doh/metrics/noop"
	prometheusmetrics "github.com/qdm12/dns/v2/pkg/doh/metrics/prometheus"
	"github.com/qdm12/dns/v2/pkg/metrics/prometheus"
	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/qdm12/dns/v2/pkg/server"
	"github.com/qdm12/gosettings"
)

func dohServer(userSettings config.Settings, ipv6Support bool,
	middlewares []Middleware, logger Logger, metrics DoHMetrics) (
	dohServer *server.Server, err error,
) {
	providers := provider.NewProviders()

	upstreamResolvers, err := stringsToUpstreamResolvers(providers, userSettings.DoH.UpstreamResolvers)
	if err != nil {
		return nil, fmt.Errorf("upstream resolvers: %w", err)
	}

	ipVersion := "ipv4"
	if ipv6Support {
		ipVersion = "ipv6"
	}

	dohSettings := doh.Settings{
		UpstreamResolvers: upstreamResolvers,
		Timeout:           userSettings.DoH.Timeout,
		IPVersion:         ipVersion,
		Metrics:           metrics,
	}
	dohDialer, err := doh.New(dohSettings)
	if err != nil {
		return nil, fmt.Errorf("creating DoH dialer: %w", err)
	}

	settings := server.Settings{
		ListeningAddress: gosettings.CopyPointer(userSettings.ListeningAddress),
		Dialer:           dohDialer,
		Middlewares:      toServerMiddlewares(middlewares),
		Logger:           logger,
	}
	return server.New(settings)
}

func dohMetrics(metricsType string, //nolint:ireturn
	commonPrometheus prometheus.Settings) (
	metrics DoHMetrics, err error,
) {
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

func stringsToUpstreamResolvers(providers *provider.Providers, providerNames []string) (
	providersSlice []provider.Provider, err error,
) {
	providersSlice = make([]provider.Provider, len(providerNames))
	for i, providerName := range providerNames {
		providersSlice[i], err = providers.Get(providerName)
		if err != nil {
			return nil, fmt.Errorf("provider %s: %w", providerName, err)
		}
	}
	return providersSlice, nil
}
