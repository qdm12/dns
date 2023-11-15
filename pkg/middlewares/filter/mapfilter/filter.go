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

func New(settings Settings) (filter *Filter, err error) {
	settings.SetDefaults()

	filter = &Filter{
		metrics: settings.Metrics,
	}

	err = filter.Update(settings.Update)
	if err != nil {
		return nil, err
	}

	return filter, nil
}
