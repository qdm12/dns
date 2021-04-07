package config

import (
	"errors"
	"fmt"

	"github.com/qdm12/dns/pkg/cache"
	"github.com/qdm12/golibs/params"
)

var ErrCacheMaxEntries = errors.New("invalid value for max entries in the cache")

func getCacheSettings(reader *reader) (settings cache.Settings, err error) {
	cacheTypes := cache.ListTypes()
	possibleCacheTypes := make([]string, len(cacheTypes))
	for i := range cacheTypes {
		possibleCacheTypes[i] = string(cacheTypes[i])
	}
	cacheType, err := reader.env.Inside("CACHE_TYPE",
		possibleCacheTypes, params.Default("lru"))
	if err != nil {
		return settings, fmt.Errorf("environment variable CACHE_TYPE: %w", err)
	}
	settings.Type = cache.Type(cacheType)

	switch settings.Type {
	case cache.Disabled:
	case cache.LRU:
		settings.LRU.MaxEntries, err = reader.env.Int("CACHE_LRU_MAX_ENTRIES",
			params.Default("10000"))
		if err != nil {
			return settings, fmt.Errorf("environment variable CACHE_LRU_MAX_ENTRIES: %w", err)
		} else if settings.LRU.MaxEntries < 1 {
			return settings, fmt.Errorf("%w: must be strictly positive: %d",
				ErrCacheMaxEntries, settings.LRU.MaxEntries)
		}
	}

	return settings, nil
}
