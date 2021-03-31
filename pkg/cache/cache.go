package cache

import (
	"github.com/miekg/dns"
)

type Cache interface {
	Add(request, response *dns.Msg)
	Get(request *dns.Msg) (response *dns.Msg)
}

func New(settings Settings) Cache {
	settings.setDefaults()

	switch settings.Type {
	case LRU:
		return newLRU(settings.MaxEntries, settings.TTL)
	case NOOP:
		return newNoop()
	default:
		// Coding error
		panic("unknown cache type: " + settings.Type)
	}
}
