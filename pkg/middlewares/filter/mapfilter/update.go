package mapfilter

import (
	"github.com/qdm12/dns/v2/pkg/middlewares/filter/update"
)

func (m *Filter) Update(settings update.Settings) (err error) {
	m.updateLock.Lock()
	defer m.updateLock.Unlock()

	m.fqdnHostnames = make(map[string]struct{}, len(settings.FqdnHostnames))
	for _, fqdnHostname := range settings.FqdnHostnames {
		m.fqdnHostnames[fqdnHostname] = struct{}{}
	}

	ipv4Count := 0
	for _, ip := range settings.IPs {
		if ip.Is4() {
			ipv4Count++
		}
	}

	m.ipv4 = make(map[[4]byte]struct{}, ipv4Count)
	m.ipv6 = make(map[[16]byte]struct{}, len(settings.IPs)-ipv4Count)
	for _, ip := range settings.IPs {
		if ip.Is4() {
			m.ipv4[ip.As4()] = struct{}{}
		} else {
			m.ipv6[ip.As16()] = struct{}{}
		}
	}

	m.ipPrefixes = settings.IPPrefixes

	m.metrics.SetBlockedHostnames(len(m.fqdnHostnames))
	m.metrics.SetBlockedIPs(len(m.ipv4) + len(m.ipv6))
	m.metrics.SetBlockedIPPrefixes(len(m.ipPrefixes))

	return nil
}
