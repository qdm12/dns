package cache

import (
	"strings"

	"github.com/qdm12/dns/pkg/cache/lru"
)

type Settings struct {
	Type Type
	LRU  lru.Settings
}

func (s *Settings) SetDefaults() {
	if string(s.Type) == "" {
		s.Type = Disabled
	}

	switch s.Type {
	case Disabled:
	case LRU:
		s.LRU.SetDefaults()
	}
}

func (s *Settings) String() string {
	const (
		subSection = " |--"
		indent     = "    " // used if lines already contain the subSection
	)
	return strings.Join(s.Lines(indent, subSection), "\n")
}

func (s *Settings) Lines(indent, subSection string) (lines []string) {
	lines = append(lines, subSection+"Type: "+string(s.Type))

	switch s.Type {
	case LRU:
		lruLines := s.LRU.Lines(indent, subSection)
		lines = append(lines, lruLines...)
	case Disabled:
	default:
		lines = append(lines, subSection+"MISSING CODE PATH, PLEASE ADD ME!!")
	}

	return lines
}
