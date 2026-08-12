package mapfilter

import (
	"net"
	"net/netip"
	"strings"

	"github.com/miekg/dns"
)

func (m *Filter) FilterResponse(response *dns.Msg) (blocked bool) {
	m.updateLock.RLock()
	defer m.updateLock.RUnlock()

	// Note the response contains the first question of
	// the request.
	nameIsLocal := false
	nameCanBeRebinded := false
	if len(response.Question) == 1 {
		nameIsLocal = m.localChecker.IsFQDNLocal(response.Question[0].Name)
		_, nameCanBeRebinded = m.allowRebindNames[response.Question[0].Name]
		if !nameCanBeRebinded && len(m.allowRebindParents) > 0 {
			labels := dns.SplitDomainName(response.Question[0].Name)
			for i := len(labels) - 1; i >= 0; i-- {
				parent := dns.Fqdn(strings.Join(labels[i:], "."))
				if _, ok := m.allowRebindParents[parent]; ok {
					nameCanBeRebinded = true
					break
				}
			}
		}
	}

	for _, rr := range response.Answer {
		// only filter A and AAAA responses for now
		rrType := rr.Header().Rrtype
		var blockedReason string
		switch rrType {
		case dns.TypeA:
			record := rr.(*dns.A) //nolint:forcetypeassert
			blockedReason = m.isIPBlocked(record.A, nameIsLocal, nameCanBeRebinded)
		case dns.TypeAAAA:
			record := rr.(*dns.AAAA) //nolint:forcetypeassert
			blockedReason = m.isIPBlocked(record.AAAA, nameIsLocal, nameCanBeRebinded)
		}

		if blockedReason != "" {
			m.metrics.IPsFilteredInc(dns.TypeToString[rrType])
			m.logger.Log("response blocked for " + rr.Header().Name + " because " + blockedReason)
			return true
		}
	}

	return false
}

func (m *Filter) isIPBlocked(ip net.IP,
	nameIsLocal, nameCanBeRebinded bool,
) (blockedReason string) {
	var netIP netip.Addr
	if ip.To4() != nil {
		ipBytes := [4]byte(ip.To4())
		_, blocked := m.ipv4[ipBytes]
		if blocked {
			return ip.String() + " is a blocked IPv4 address"
		}
		netIP = netip.AddrFrom4(ipBytes)
	} else {
		ipBytes := [16]byte(ip.To16())
		_, blocked := m.ipv6[ipBytes]
		if blocked {
			return ip.String() + " is a blocked IPv6 address"
		}
		netIP = netip.AddrFrom16(ipBytes)
	}

	// Only run the rebinding protection on non-local question names
	// which are also not exempt from rebinding protection.
	if !nameIsLocal && !nameCanBeRebinded {
		for _, ipPrefix := range m.privateIPPrefixes {
			if ipPrefix.Contains(netIP) {
				return netIP.String() + " is private and the question name is not local nor exempt from rebinding protection"
			}
		}
	}

	for _, ipPrefix := range m.ipPrefixes {
		if ipPrefix.Contains(netIP) {
			return netIP.String() + " belongs to the blocked IP prefix " + ipPrefix.String()
		}
	}
	return ""
}
