package cache

import (
	"strings"

	"github.com/qdm12/dns/pkg/cache/lru"
	"github.com/qdm12/dns/pkg/cache/noop"
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
	const (
		subSection = " |--"
		indent     = "    " // used if lines already contain the subSection
	)
	return strings.Join(s.Lines(indent, subSection), "\n")
}

func (s *Settings) Lines(indent, subSection string) (lines []string) {
	switch s.Type {
	case lru.CacheType:
		lruLines := s.LRU.Lines(indent, subSection)
		lines = append(lines, lruLines...)
	case noop.CacheType:
	default:
		lines = append(lines, subSection+"MISSING CODE PATH, PLEASE ADD ME!!")
	}

	return lines
}
