package mapfilter

import (
	"net/netip"
	"sync"
)

type Filter struct {
	fqdnHostnames map[string]struct{}
	ips           map[netip.Addr]struct{}
	ipPrefixes    []netip.Prefix
	metrics       Metrics
	updateLock    sync.RWMutex
}

func New(settings Settings) *Filter {
	settings.SetDefaults()

	filter := &Filter{
		metrics: settings.Metrics,
	}

	filter.Update(settings.Update)

	return filter
}
