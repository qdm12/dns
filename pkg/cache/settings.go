package cache

import (
	"github.com/qdm12/dns/pkg/cache/lru"
	"github.com/qdm12/dns/pkg/cache/noop"
	"github.com/qdm12/gotree"
)

type Settings struct {
	Type string
	LRU  lru.Settings
	Noop noop.Settings
}

func (s *Settings) SetDefaults() {
	if s.Type == "" {
		s.Type = noop.CacheType
	}

	// cache implementations defaults set by their constructor
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
		// TODO use ToLinesNode if it exists
		return nil
	}
}
