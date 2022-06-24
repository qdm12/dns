package setup

import (
	"fmt"

	"github.com/miekg/dns"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/v2/internal/config/settings"
	"github.com/qdm12/dns/v2/pkg/cache/lru"
	"github.com/qdm12/dns/v2/pkg/cache/metrics"
	noopmetrics "github.com/qdm12/dns/v2/pkg/cache/metrics/noop"
	prommetrics "github.com/qdm12/dns/v2/pkg/cache/metrics/prometheus"
	"github.com/qdm12/dns/v2/pkg/cache/noop"
	promcommon "github.com/qdm12/dns/v2/pkg/metrics/prometheus"
)

type Cache interface {
	Add(request, response *dns.Msg)
	Get(request *dns.Msg) (response *dns.Msg)
}

func BuildCache(userSettings settings.Cache, //nolint:ireturn
	metrics metrics.Interface) (cache Cache) {
	switch userSettings.Type {
	case noop.CacheType:
		return noop.New(noop.Settings{Metrics: metrics})
	case lru.CacheType:
		return lru.New(lru.Settings{
			MaxEntries: userSettings.LRU.MaxEntries,
			Metrics:    metrics,
		})
	default:
		panic(fmt.Sprintf("unknown cache type: %s", userSettings.Type))
	}
}

func CacheMetrics(userSettings settings.Metrics, //nolint:ireturn
	registry prometheus.Registerer) (
	metrics metrics.Interface, err error) {
	switch userSettings.Type {
	case noopString:
		return noopmetrics.New(), nil
	case prometheusString:
		settings := prommetrics.Settings{
			Prometheus: promcommon.Settings{
				Registry: registry,
				Prefix:   "",
			},
		}
		metrics, err = prommetrics.New(settings)
		if err != nil {
			return nil, fmt.Errorf("setup Prometheus metrics: %w", err)
		}
	default:
		panic(fmt.Sprintf("unknown metrics type: %s", userSettings.Type))
	}

	return metrics, nil
}