package mapfilter

import (
	"net/netip"

	"github.com/qdm12/dns/v2/pkg/middlewares/filter/update"
)

func (m *Filter) Update(settings update.Settings) {
	m.updateLock.Lock()
	defer m.updateLock.Unlock()

	m.fqdnHostnames = make(map[string]struct{}, len(settings.FqdnHostnames))
	for _, fqdnHostname := range settings.FqdnHostnames {
		m.fqdnHostnames[fqdnHostname] = struct{}{}
	}

	m.ips = make(map[netip.Addr]struct{}, len(settings.IPs))
	for _, ip := range settings.IPs {
		m.ips[ip] = struct{}{}
	}

	m.ipPrefixes = settings.IPPrefixes

	m.metrics.SetBlockedHostnames(len(m.fqdnHostnames))
	m.metrics.SetBlockedIPs(len(m.ips))
	m.metrics.SetBlockedIPPrefixes(len(m.ips))
}
