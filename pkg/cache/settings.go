package cache

import (
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
