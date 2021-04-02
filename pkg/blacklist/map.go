package blacklist

import (
	"github.com/miekg/dns"
)

type mapBased struct {
	fqdnHostnames map[string]struct{}
	ips           map[string]struct{}
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
