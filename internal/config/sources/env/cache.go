package env

import (
	"errors"
	"fmt"
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

func getLRUCacheMaxEntries() (maxEntries int, err error) {
	s := os.Getenv("CACHE_LRU_MAX_ENTRIES")
	if s == "" {
		return 0, nil
	}

	maxEntries, err = strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("environment variable CACHE_LRU_MAX_ENTRIES: %w", err)
	} else if maxEntries < 1 {
		return 0, fmt.Errorf("%w: must be strictly positive: %d",
			ErrCacheLRUMaxEntries, maxEntries)
	}
	return maxEntries, nil
}
