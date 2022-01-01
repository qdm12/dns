package cache

import (
	"errors"
	"fmt"

	"github.com/qdm12/dns/pkg/cache/lru"
	"github.com/qdm12/dns/pkg/cache/noop"
	"github.com/qdm12/gotree"
)

type Settings struct { // TODO move to internal
	Type string
	LRU  lru.Settings
	Noop noop.Settings
}

func (s *Settings) SetDefaults() {
	if s.Type == "" {
		s.Type = noop.CacheType
	}

	switch s.Type {
	case noop.CacheType:
		s.Noop.SetDefaults()
	case lru.CacheType:
		s.LRU.SetDefaults()
	}
}

var (
	ErrCacheTypeNotValid = errors.New("cache type is not valid")
)

func (s Settings) Validate() (err error) {
	switch s.Type {
	case noop.CacheType, lru.CacheType:
	default:
		return fmt.Errorf("%w: %s", ErrCacheTypeNotValid, s.Type)
	}

	err = s.LRU.Validate()
	if err != nil {
		return fmt.Errorf("failed validating LRU cache settings: %w", err)
	}

	err = s.Noop.Validate()
	if err != nil {
		return fmt.Errorf("failed validating Noop cache settings: %w", err)
	}

	return nil
}

func (s *Settings) String() string {
	return s.ToLinesNode().String()
}

func (s *Settings) ToLinesNode() (node *gotree.Node) {
	switch s.Type {
	case lru.CacheType:
		return s.LRU.ToLinesNode()
	case noop.CacheType:
		return nil
	default:
		return nil
	}
}
