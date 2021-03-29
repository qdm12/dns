package cache

import "time"

type Option func(s *settings)

func MaxEntries(maxEntries int) Option {
	return func(s *settings) {
		s.maxEntries = maxEntries
	}
}

func TTL(ttl time.Duration) Option {
	return func(s *settings) {
		s.ttl = ttl
	}
}
