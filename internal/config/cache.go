package config

import (
	"errors"
	"fmt"

	"github.com/qdm12/dns/pkg/cache"
	"github.com/qdm12/dns/pkg/cache/lru"
	"github.com/qdm12/dns/pkg/cache/noop"
	"github.com/qdm12/golibs/params"
	"github.com/qdm12/gotree"
)

func (settings *Settings) PatchCache(cache cache.Interface) {
	settings.DoT.Cache = cache
	settings.DoH.Cache = cache
}

var errCacheMaxEntries = errors.New("invalid value for max entries in the cache")

type Cache struct {
	Type string
	LRU  lru.Settings
	Noop noop.Settings
}

func (c *Cache) SetDefaults() {
	if c.Type == "" {
		c.Type = noop.CacheType
	}

	switch c.Type {
	case noop.CacheType:
		c.Noop.SetDefaults()
	case lru.CacheType:
		c.LRU.SetDefaults()
	}
}

var (
	ErrCacheTypeNotValid = errors.New("cache type is not valid")
)

func (c Cache) Validate() (err error) {
	switch c.Type {
	case noop.CacheType, lru.CacheType:
	default:
		return fmt.Errorf("%w: %s", ErrCacheTypeNotValid, c.Type)
	}

	err = c.LRU.Validate()
	if err != nil {
		return fmt.Errorf("failed validating LRU cache settings: %w", err)
	}

	err = c.Noop.Validate()
	if err != nil {
		return fmt.Errorf("failed validating Noop cache settings: %w", err)
	}

	return nil
}

func (c *Cache) String() string {
	return c.ToLinesNode().String()
}

func (c *Cache) ToLinesNode() (node *gotree.Node) {
	switch c.Type {
	case lru.CacheType:
		return c.LRU.ToLinesNode()
	case noop.CacheType:
		return nil
	default:
		return nil
	}
}

func getCacheSettings(reader *Reader) (settings Cache, err error) {
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
