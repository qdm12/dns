package config

import (
	"errors"
	"fmt"

	"github.com/qdm12/dns/pkg/cache"
	"github.com/qdm12/dns/pkg/cache/lru"
	"github.com/qdm12/dns/pkg/cache/noop"
	"github.com/qdm12/golibs/params"
)

func (settings *Settings) PatchCache(cache cache.Interface) {
	settings.DoT.Cache = cache
	settings.DoH.Cache = cache
}

var errCacheMaxEntries = errors.New("invalid value for max entries in the cache")

func getCacheSettings(reader *Reader) (settings cache.Settings, err error) {
	settings.Type, err = reader.env.Inside("CACHE_TYPE",
		[]string{lru.CacheType, noop.CacheType},
		params.Default(lru.CacheType))
	if err != nil {
		return settings, fmt.Errorf("environment variable CACHE_TYPE: %w", err)
	}

	settings.LRU.MaxEntries, err = reader.env.Int(
		"CACHE_LRU_MAX_ENTRIES", params.Default("10000"))
	if err != nil {
		return settings, fmt.Errorf("environment variable CACHE_LRU_MAX_ENTRIES: %w", err)
	} else if settings.LRU.MaxEntries < 1 {
		return settings, fmt.Errorf("%w: must be strictly positive: %d",
			errCacheMaxEntries, settings.LRU.MaxEntries)
	}

	return settings, nil
}
