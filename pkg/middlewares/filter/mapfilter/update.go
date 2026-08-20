package mapfilter

import (
	"fmt"
	"net/netip"

	"github.com/qdm12/dns/v2/internal/local"
	"github.com/qdm12/dns/v2/pkg/middlewares/filter/update"
)

func (m *Filter) Update(settings update.Settings) (err error) {
	m.updateLock.Lock()
	defer m.updateLock.Unlock()

	m.localChecker = local.New(settings.PublicFQDNsAsLocal)
	m.blockHostnames(settings.FqdnHostnames)
	m.blockIPs(settings.IPs)
	m.blockIPPrefixes(settings.IPPrefixes)
	m.setRebindingProtectionExempt(settings.FqdnExemptFromRebindingProtection)
	m.setRebindingProtectionExemptParents(settings.ParentsExemptFromRebindingProtection)
	m.metrics.SetFqdnExemptFromRebindingProtection(len(m.allowRebindNames) + len(m.allowRebindParents))

	m.logger.Log(fmt.Sprintf("filter updated: %d hostnames, %d IPs, %d IP prefixes blocked",
		len(m.fqdnHostnames), len(m.ipv4)+len(m.ipv6), len(m.ipPrefixes)))

	return nil
}

func (m *Filter) blockHostnames(fqdnHostnames []string) {
	if fqdnHostnames == nil {
		return
	}
	m.fqdnHostnames = make(map[string]struct{}, len(fqdnHostnames))
	for _, fqdnHostname := range fqdnHostnames {
		m.fqdnHostnames[fqdnHostname] = struct{}{}
	}
	m.metrics.SetBlockedHostnames(len(m.fqdnHostnames))
}

func (m *Filter) blockIPs(ips []netip.Addr) {
	if ips == nil {
		return
	}

	ipv4Count := 0
	for _, ip := range ips {
		if ip.Is4() {
			ipv4Count++
		}
	}

	m.ipv4 = make(map[[4]byte]struct{}, ipv4Count)
	m.ipv6 = make(map[[16]byte]struct{}, len(ips)-ipv4Count)
	for _, ip := range ips {
		if ip.Is4() {
			m.ipv4[ip.As4()] = struct{}{}
		} else {
			m.ipv6[ip.As16()] = struct{}{}
		}
	}
	m.metrics.SetBlockedIPs(len(m.ipv4) + len(m.ipv6))
}

func (m *Filter) blockIPPrefixes(ipPrefixes []netip.Prefix) {
	if ipPrefixes == nil {
		return
	}
	m.ipPrefixes = ipPrefixes
	m.metrics.SetBlockedIPPrefixes(len(m.ipPrefixes))
}

func (m *Filter) setRebindingProtectionExempt(hostnames []string) {
	if hostnames == nil {
		return
	}
	m.allowRebindNames = make(map[string]struct{}, len(hostnames))
	for _, name := range hostnames {
		m.allowRebindNames[name] = struct{}{}
	}
}

func (m *Filter) setRebindingProtectionExemptParents(hostnames []string) {
	if hostnames == nil {
		return
	}
	m.allowRebindParents = make(map[string]struct{}, len(hostnames))
	for _, name := range hostnames {
		m.allowRebindParents[name] = struct{}{}
	}
	m.metrics.SetFqdnExemptFromRebindingProtection(len(m.allowRebindNames) + len(m.allowRebindParents))
}
