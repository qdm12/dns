package cache

import (
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
