package mapfilter

import (
	"net"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/pkg/filter/metrics"
	"inet.af/netaddr"
)

type Filter struct {
	fqdnHostnames map[string]struct{}
	ips           map[netaddr.IP]struct{}
	ipPrefixes    []netaddr.IPPrefix
	metrics       metrics.Interface
}

func New(settings Settings) *Filter {
	settings.setDefaults()
	metrics := settings.Metrics

	fqdnHostnamesSet := make(map[string]struct{}, len(settings.FqdnHostnames))
	for _, fqdnHostname := range settings.FqdnHostnames {
		fqdnHostnamesSet[fqdnHostname] = struct{}{}
	}
	metrics.SetBlockedHostnames(len(fqdnHostnamesSet))

	ipsSet := make(map[netaddr.IP]struct{}, len(settings.IPs))
	for _, ip := range settings.IPs {
		ipsSet[ip] = struct{}{}
	}
	metrics.SetBlockedIPs(len(ipsSet))

	metrics.SetBlockedIPPrefixes(len(settings.IPPrefixes))

	return &Filter{
		fqdnHostnames: fqdnHostnamesSet,
		ips:           ipsSet,
		ipPrefixes:    settings.IPPrefixes,
		metrics:       metrics,
	}
}

func (m *Filter) FilterRequest(request *dns.Msg) (blocked bool) {
	for _, question := range request.Question {
		fqdnHostname := question.Name
		_, blocked = m.fqdnHostnames[fqdnHostname]
		if blocked {
			class := dns.ClassToString[question.Qclass]
			qType := dns.TypeToString[question.Qtype]
			m.metrics.HostnamesFilteredInc(class, qType)
			return blocked
		}
	}
	return false
}

func (m *Filter) FilterResponse(response *dns.Msg) (blocked bool) {
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
