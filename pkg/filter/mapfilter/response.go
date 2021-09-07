package mapfilter

import (
	"net"

	"github.com/miekg/dns"
	"inet.af/netaddr"
)

func (m *Filter) FilterResponse(response *dns.Msg) (blocked bool) {
	m.updateLock.RLock()
	defer m.updateLock.RUnlock()

	for _, rr := range response.Answer {
		// only filter A and AAAA responses for now
		rrType := rr.Header().Rrtype
		switch rrType {
		case dns.TypeA:
			record := rr.(*dns.A)
			blocked = m.isIPBlocked(record.A)
		case dns.TypeAAAA:
			record := rr.(*dns.AAAA)
			blocked = m.isIPBlocked(record.AAAA)
		}

		if blocked {
			m.metrics.IPsFilteredInc(dns.TypeToString[rrType])
			return true
		}
	}

	return false
}

func (m *Filter) isIPBlocked(ip net.IP) (blocked bool) {
	netaddrIP, ok := netaddr.FromStdIP(ip)
	if !ok {
		return true
	}

	if _, blocked := m.ips[netaddrIP]; blocked {
		return blocked
	}

	for _, ipPrefix := range m.ipPrefixes {
		if ipPrefix.Contains(netaddrIP) {
			return true
		}
	}
	return false
}
