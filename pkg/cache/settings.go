package cache

import (
	"strconv"
	"strings"
	"time"
)

type Settings struct {
	Type       Type
	MaxEntries int
	TTL        time.Duration
}

func (s *Settings) setDefaults() {
	if string(s.Type) == "" {
		s.Type = LRU
	}

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
	if s.Type == NOOP {
		return []string{subSection + "Type: No-op (disabled)"}
	}

	lines = append(lines, "Type: "+string(s.Type))
	lines = append(lines, "Max entries: "+strconv.Itoa(s.MaxEntries))
	lines = append(lines, "Entry TTL: "+s.TTL.String())

	return lines
}
