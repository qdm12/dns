package blacklist

import (
	"net"

	"github.com/miekg/dns"
)

type mapBased struct {
	fqdnHostnames map[string]struct{}
	ips           map[string]struct{}
}

func NewMap(fqdnHostnames []string, ips []net.IP) BlackLister {
	fqdnHostnamesSet := make(map[string]struct{}, len(fqdnHostnames))
	for _, fqdnHostname := range fqdnHostnames {
		fqdnHostnamesSet[fqdnHostname] = struct{}{}
	}

	ipsSet := make(map[string]struct{}, len(ips))
	for _, ip := range ips {
		ipsSet[ip.String()] = struct{}{}
	}

	return &mapBased{
		fqdnHostnames: fqdnHostnamesSet,
		ips:           ipsSet,
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
			ipStr := record.A.String()
			if _, blocked := m.ips[ipStr]; blocked {
				return blocked
			}
		case dns.TypeAAAA:
			record := rr.(*dns.AAAA)
			ipStr := record.AAAA.String()
			if _, blocked := m.ips[ipStr]; blocked {
				return blocked
			}
		}
	}
	return false
}
