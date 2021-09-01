package metrics

import (
	"github.com/qdm12/dns/pkg/cache/metrics/noop"
	"github.com/qdm12/dns/pkg/cache/metrics/prometheus"
)

var (
	_ Interface = (*prometheus.Metrics)(nil)
	_ Interface = (*noop.Metrics)(nil)
)

//go:generate mockgen -destination=mock_$GOPACKAGE/$GOFILE . Interface

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
