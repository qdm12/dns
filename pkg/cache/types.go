package cache

import (
	"errors"
	"fmt"
	"strings"
)

type Type string

const (
	LRU      Type = "lru"
	Disabled Type = "disabled"
)

func ListTypes() (types []Type) {
	return []Type{
		LRU,
		Disabled,
	}
}

var ErrParseCacheType = errors.New("cannot parse cache type")

func ParseCacheType(s string) (cacheType Type, err error) {
	for _, T := range ListTypes() {
		if strings.EqualFold(string(T), s) {
			return T, nil
		}
	}
	return "", fmt.Errorf("%w: %q is unknown", ErrParseCacheType, s)
}
