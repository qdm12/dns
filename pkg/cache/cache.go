package cache

import (
	"github.com/miekg/dns"
)

type Cache interface {
	Add(request, response *dns.Msg)
	Get(request *dns.Msg) (response *dns.Msg)
}

func New(cacheType Type, options ...Option) Cache {
	settings := defaultSettings()
	for _, option := range options {
		option(&settings)
	}

	switch cacheType {
	case LRU:
		return newLRU(settings.maxEntries, settings.ttl)
	case NOOP:
		return newNoop()
	default:
		// Coding error
		panic("unknown cache type")
	}
}
