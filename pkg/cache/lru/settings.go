package lru

import (
	"strconv"
	"strings"
	"time"
)

type Settings struct {
	MaxEntries int
	TTL        time.Duration
}

func (s *Settings) SetDefaults() {
	if s.MaxEntries == 0 {
		s.MaxEntries = 10e4
	}

	if s.TTL == 0 {
		s.TTL = time.Hour
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
	lines = append(lines, subSection+"Entry TTL: "+s.TTL.String())
	return lines
}
