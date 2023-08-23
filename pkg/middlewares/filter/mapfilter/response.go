package mapfilter

import (
	"net"
	"net/netip"

	"github.com/miekg/dns"
)

func (m *Filter) FilterResponse(response *dns.Msg) (blocked bool) {
	m.updateLock.RLock()
	defer m.updateLock.RUnlock()

	for _, rr := range response.Answer {
		// only filter A and AAAA responses for now
		rrType := rr.Header().Rrtype
		switch rrType {
		case dns.TypeA:
			record := rr.(*dns.A) //nolint:forcetypeassert
			blocked = m.isIPBlocked(record.A)
		case dns.TypeAAAA:
			record := rr.(*dns.AAAA) //nolint:forcetypeassert
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
	var netIP netip.Addr
	if ip.To4() != nil {
		netIP = netip.AddrFrom4([4]byte(ip.To4()))
	} else {
		netIP = netip.AddrFrom16([16]byte(ip.To16()))
	}

	if _, blocked := m.ips[netIP]; blocked {
		return blocked
	}

	for _, ipPrefix := range m.ipPrefixes {
		if ipPrefix.Contains(netIP) {
			return true
		}
	}
	return false
}
