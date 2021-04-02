package blacklist

import (
	"net"

	"github.com/miekg/dns"
)

type mapBased struct {
	fqdnHostnames map[string]struct{}
	ips           map[string]struct{}
	ipNets        []*net.IPNet
}

func NewMap(settings Settings) BlackLister {
	fqdnHostnamesSet := make(map[string]struct{}, len(settings.FqdnHostnames))
	for _, fqdnHostname := range settings.FqdnHostnames {
		fqdnHostnamesSet[fqdnHostname] = struct{}{}
	}

	ipsSet := make(map[string]struct{}, len(settings.IPs))
	for _, ip := range settings.IPs {
		ipsSet[ip.String()] = struct{}{}
	}

	return &mapBased{
		fqdnHostnames: fqdnHostnamesSet,
		ips:           ipsSet,
		ipNets:        settings.IPNets,
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
	ipStr := ip.String()
	if _, blocked := m.ips[ipStr]; blocked {
		return blocked
	}
	for _, ipNet := range m.ipNets {
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}
