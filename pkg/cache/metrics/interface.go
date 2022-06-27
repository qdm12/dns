// Package metrics defines an interface valid for all caches.
package metrics

import (
	"fmt"

	"github.com/qdm12/dns/v2/pkg/cache/metrics/noop"
	"github.com/qdm12/dns/v2/pkg/cache/metrics/prometheus"
)

var (
	_ Interface = (*prometheus.Metrics)(nil)
	_ Interface = (*noop.Metrics)(nil)
)

type Interface interface {
	SetCacheType(cacheType string)
	CacheInsertInc()
	CacheRemoveInc()
	CacheMoveInc()
	CacheGetEmptyInc()
	CacheInsertEmptyInc()
	CacheHitInc()
	CacheMissInc()
	CacheExpiredInc()
	CacheMaxEntriesSet(maxEntries int)
}

type Settings struct {
	Type       string
	Prometheus prometheus.Settings
}

func Metrics(settings Settings) (metrics Interface, err error) { //nolint:ireturn
	switch settings.Type {
	case "noop":
		return noop.New(), nil
	case "prometheus":
		metrics, err = prometheus.New(settings.Prometheus)
		if err != nil {
			return nil, fmt.Errorf("setting up Prometheus metrics: %w", err)
		}
	default:
		panic(fmt.Sprintf("unknown metrics type: %s", settings.Type))
	}

	return metrics, nil
}
