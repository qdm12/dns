package setup

import (
	"fmt"

	"github.com/qdm12/dns/v2/internal/config"
	"github.com/qdm12/dns/v2/pkg/dot"
	noopmetrics "github.com/qdm12/dns/v2/pkg/dot/metrics/noop"
	prometheusmetrics "github.com/qdm12/dns/v2/pkg/dot/metrics/prometheus"
	"github.com/qdm12/dns/v2/pkg/metrics/prometheus"
	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/qdm12/dns/v2/pkg/server"
	"github.com/qdm12/gosettings"
)

func dotServer(userSettings config.Settings, ipv6Support bool,
	middlewares []Middleware, logger Logger, metrics DoTMetrics,
	poolMetrics PoolMetrics) (
	dotServer *server.Server, err error,
) {
	providers := provider.NewProviders()

	upstreamResolvers, err := stringsToUpstreamResolvers(providers, userSettings.DoT.UpstreamResolvers)
	if err != nil {
		return nil, fmt.Errorf("upstream resolvers: %w", err)
	}

	ipVersion := "ipv4"
	if ipv6Support {
		ipVersion = "ipv6"
	}

	dialerSettings := dot.Settings{
		UpstreamResolvers: upstreamResolvers,
		IPVersion:         ipVersion,
		Metrics:           metrics,
	}
	dialer, err := dot.New(dialerSettings)
	if err != nil {
		return nil, fmt.Errorf("creating DoT dialer: %w", err)
	}

	serverSettings := server.Settings{
		ListeningAddress: gosettings.CopyPointer(userSettings.ListeningAddress),
		Dialer:           dialer,
		PoolMetrics:      poolMetrics,
		Middlewares:      toServerMiddlewares(middlewares),
		Logger:           logger,
	}
	return server.New(serverSettings)
}

func dotMetrics(metricsType string, //nolint:ireturn
	commonPrometheus prometheus.Settings) (
	metrics DoTMetrics, err error,
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
