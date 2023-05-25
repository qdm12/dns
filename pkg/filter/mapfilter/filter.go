package mapfilter

import (
	"net/netip"
	"sync"

	"github.com/qdm12/dns/v2/pkg/filter/metrics"
)

type Filter struct {
	fqdnHostnames map[string]struct{}
	ips           map[netip.Addr]struct{}
	ipPrefixes    []netip.Prefix
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
