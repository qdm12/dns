package setup

import (
	"fmt"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/internal/config/settings"
	promcommon "github.com/qdm12/dns/v2/pkg/metrics/prometheus"
	"github.com/qdm12/dns/v2/pkg/middlewares/cache/lru"
	noopmetrics "github.com/qdm12/dns/v2/pkg/middlewares/cache/metrics/noop"
	prommetrics "github.com/qdm12/dns/v2/pkg/middlewares/cache/metrics/prometheus"
	"github.com/qdm12/dns/v2/pkg/middlewares/cache/noop"
)

type Cache interface {
	Add(request, response *dns.Msg)
	Get(request *dns.Msg) (response *dns.Msg)
}

func BuildCache(userSettings settings.Cache, //nolint:ireturn
	metrics CacheMetrics) (cache Cache, err error) {
	switch userSettings.Type {
	case noop.CacheType:
		return noop.New(noop.Settings{Metrics: metrics}), nil
	case lru.CacheType:
		return lru.New(lru.Settings{
			MaxEntries: userSettings.LRU.MaxEntries,
			Metrics:    metrics,
		})
	default:
		panic(fmt.Sprintf("unknown cache type: %s", userSettings.Type))
	}
}

func BuildCacheMetrics(userSettings settings.Metrics, //nolint:ireturn
	registry PrometheusRegistry) (
	metrics CacheMetrics, err error) {
	switch userSettings.Type {
	case noopString:
		return noopmetrics.New(), nil
	case prometheusString:
		settings := prommetrics.Settings{
			Prometheus: promcommon.Settings{
				Registry: registry,
				Prefix:   *userSettings.Prometheus.Subsystem,
			},
		}
		metrics, err = prommetrics.New(settings)
		if err != nil {
			return nil, fmt.Errorf("setting up Prometheus metrics: %w", err)
		}
	default:
		panic(fmt.Sprintf("unknown metrics type: %s", userSettings.Type))
	}

	return metrics, nil
}
