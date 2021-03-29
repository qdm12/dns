package cache

import (
	"time"
)

type settings struct {
	maxEntries int
	ttl        time.Duration
}

func defaultSettings() (settings settings) {
	settings.maxEntries = 10e4
	settings.ttl = time.Hour
	return settings
}
