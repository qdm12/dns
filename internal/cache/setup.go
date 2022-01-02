package cache

import (
	"github.com/qdm12/dns/v2/internal/config"
	"github.com/qdm12/dns/v2/pkg/cache/lru"
	"github.com/qdm12/dns/v2/pkg/cache/noop"
)

func Setup(settings *config.Settings) {
	if settings.Cache.Type == lru.CacheType {
		cache := lru.New(lru.Settings{
			MaxEntries: settings.Cache.LRU.MaxEntries,
			Metrics:    settings.Cache.LRU.Metrics,
		})
		settings.PatchCache(cache)
		return
	}

	// noop
	cache := noop.New(noop.Settings{
		Metrics: settings.Cache.Noop.Metrics,
	})
	settings.PatchCache(cache)
}
