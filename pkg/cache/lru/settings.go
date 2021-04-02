package lru

import (
	"strconv"
	"strings"
)

type Settings struct {
	MaxEntries int
}

func (s *Settings) SetDefaults() {
	if s.MaxEntries == 0 {
		s.MaxEntries = 10e4
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
	lines = append(lines, subSection+"Max entries: "+strconv.Itoa(s.MaxEntries))
	return lines
}
