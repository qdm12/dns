package env

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/qdm12/dns/v2/internal/config/settings"
)

func readCache() (settings settings.Cache, err error) {
	settings.Type = strings.ToLower(os.Getenv("CACHE_TYPE"))

	settings.LRU.MaxEntries, err = getLRUCacheMaxEntries()
	if err != nil {
		return settings, fmt.Errorf("LRU max entries: %w", err)
	}

	return settings, nil
}

var ErrCacheLRUMaxEntries = errors.New("invalid value for max entries of the LRU cache")

func getLRUCacheMaxEntries() (maxEntries uint, err error) {
	s := os.Getenv("CACHE_LRU_MAX_ENTRIES")
	if s == "" {
		return 0, nil
	}

	const base, bits = 10, 64
	maxEntriesUint64, err := strconv.ParseUint(s, base, bits)
	switch {
	case err != nil:
		return 0, fmt.Errorf("environment variable CACHE_LRU_MAX_ENTRIES: %w", err)
	case maxEntries == 0:
		return 0, fmt.Errorf("%w: cannot be zero", ErrCacheLRUMaxEntries)
	case maxEntries > math.MaxInt:
		// down the call stack, maxEntries is converted to int
		// for a map size, and the max int depends on the platform.
		return 0, fmt.Errorf("%w: %s must be less than %d",
			ErrCacheLRUMaxEntries, s, math.MaxInt)
	}

	maxEntries = uint(maxEntriesUint64)
	return maxEntries, nil
}
