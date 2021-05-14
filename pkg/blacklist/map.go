package blacklist

import (
	"net"

	"github.com/miekg/dns"
	"inet.af/netaddr"
)

type mapBased struct {
	fqdnHostnames map[string]struct{}
	ips           map[netaddr.IP]struct{}
	ipPrefixes    []netaddr.IPPrefix
}

func NewMap(settings Settings) BlackLister {
	fqdnHostnamesSet := make(map[string]struct{}, len(settings.FqdnHostnames))
	for _, fqdnHostname := range settings.FqdnHostnames {
		fqdnHostnamesSet[fqdnHostname] = struct{}{}
	}

	ipsSet := make(map[netaddr.IP]struct{}, len(settings.IPs))
	for _, ip := range settings.IPs {
		ipsSet[ip] = struct{}{}
	}

	return &mapBased{
		fqdnHostnames: fqdnHostnamesSet,
		ips:           ipsSet,
		ipPrefixes:    settings.IPPrefixes,
	}
}

func (m *mapBased) FilterRequest(request *dns.Msg) (blocked bool) {
	for _, question := range request.Question {
		fqdnHostname := question.Name
		if _, blocked := m.fqdnHostnames[fqdnHostname]; blocked {
			return blocked
		}
	}
	return false
}

func (m *mapBased) FilterResponse(response *dns.Msg) (blocked bool) {
	for _, rr := range response.Answer {
		// only filter A and AAAA responses for now
		switch rr.Header().Rrtype {
		case dns.TypeA:
			record := rr.(*dns.A)
			if blocked := m.isIPBlocked(record.A); blocked {
				return blocked
			}
		case dns.TypeAAAA:
			record := rr.(*dns.AAAA)
			if blocked := m.isIPBlocked(record.AAAA); blocked {
				return blocked
			}
		}
	}
	return false
}

func (m *mapBased) isIPBlocked(ip net.IP) (blocked bool) {
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
