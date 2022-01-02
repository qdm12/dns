// Package metrics defines an interface valid for all caches.
package metrics

import (
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
