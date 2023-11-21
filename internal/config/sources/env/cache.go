package env

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/qdm12/dns/v2/internal/config/settings"
)

func (r *Reader) readCache() (settings settings.Cache, err error) {
	settings.Type = r.reader.String("CACHE_TYPE")

	settings.LRU.MaxEntries, err = r.getLRUCacheMaxEntries()
	if err != nil {
		return settings, fmt.Errorf("LRU max entries: %w", err)
	}

	return settings, nil
}

var ErrCacheLRUMaxEntries = errors.New("invalid value for max entries of the LRU cache")

func (r *Reader) getLRUCacheMaxEntries() (maxEntries uint, err error) {
	s := r.reader.String("CACHE_LRU_MAX_ENTRIES")
	if s == "" {
		return 0, nil
	}

	const base, bits = 10, 64
	maxEntriesUint64, err := strconv.ParseUint(s, base, bits)
	switch {
	case err != nil:
		return 0, fmt.Errorf("environment variable CACHE_LRU_MAX_ENTRIES: %w", err)
	case maxEntriesUint64 == 0:
		return 0, fmt.Errorf("%w: cannot be zero", ErrCacheLRUMaxEntries)
	case maxEntriesUint64 > math.MaxInt:
		// down the call stack, maxEntries is converted to int
		// for a map size, and the max int depends on the platform.
		return 0, fmt.Errorf("%w: %s must be less than %d",
			ErrCacheLRUMaxEntries, s, math.MaxInt)
	}

	maxEntries = uint(maxEntriesUint64)
	return maxEntries, nil
}
