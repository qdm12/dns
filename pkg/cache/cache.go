package cache

import (
	"github.com/miekg/dns"
	"github.com/qdm12/dns/pkg/cache/lru"
)

//go:generate mockgen -destination=mock_$GOPACKAGE/$GOFILE . Cache

type Cache interface {
	Add(request, response *dns.Msg)
	Get(request *dns.Msg) (response *dns.Msg)
}

// New creates a new cache object except when the cache type
// is set to Disabled. In this case it returns a nil Cache.
func New(settings Settings) Cache {
	settings.SetDefaults()
	switch settings.Type {
	case LRU:
		return lru.New(settings.LRU)
	case Disabled:
		return nil
	default: // coding error as an end user should use ParseType
		panic("unknown cache type: " + settings.Type)
	}
}
