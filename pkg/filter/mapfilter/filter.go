package mapfilter

import (
	"sync"

	"github.com/qdm12/dns/v2/pkg/filter/metrics"
	"inet.af/netaddr"
)

type Filter struct {
	fqdnHostnames map[string]struct{}
	ips           map[netaddr.IP]struct{}
	ipPrefixes    []netaddr.IPPrefix
	metrics       metrics.Interface
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
