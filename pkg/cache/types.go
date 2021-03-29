package cache

import (
	"errors"
	"fmt"
	"strings"
)

type Type string

const (
	LRU  Type = "lru"
	NOOP Type = "noop"
)

var ErrParseCacheType = errors.New("cannot parse cache type")

func ParseCacheType(s string) (cacheType Type, err error) {
	switch strings.ToLower(s) {
	case string(LRU):
		return LRU, nil
	case string(NOOP):
		return NOOP, nil
	default:
		return "", fmt.Errorf("%w: %q is unknown", ErrParseCacheType, s)
	}
}
