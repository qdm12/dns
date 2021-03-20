package dot

import (
	"net"
	"sync/atomic"

	"github.com/qdm12/dns/pkg/provider"
)

type picker struct {
	counterDoTServer *int64
	counterDoTIPv4   *int64
	counterDoTIPv6   *int64
	counterDNSServer *int64
	counterDNSIPv4   *int64
	counterDNSIPv6   *int64
}

func newPicker() *picker {
	return &picker{
		counterDoTServer: new(int64),
		counterDoTIPv4:   new(int64),
		counterDoTIPv6:   new(int64),
		counterDNSServer: new(int64),
		counterDNSIPv4:   new(int64),
		counterDNSIPv6:   new(int64),
	}
}

func (p *picker) randIndex(counter *int64, max int) int {
	index := int(atomic.AddInt64(counter, 1))
	return index % max
}

func (p *picker) DoTServer(servers []provider.DoTServer) provider.DoTServer {
	index := 0
	if nServers := len(servers); nServers > 1 {
		index = p.randIndex(p.counterDoTServer, nServers)
	}
	return servers[index]
}

func (p *picker) IP(counter *int64, ips []net.IP) net.IP {
	switch len(ips) {
	case 0:
		return nil
	case 1:
		return ips[0]
	default:
		index := p.randIndex(counter, len(ips))
		return ips[index]
	}
}

func (p *picker) DoTIP(server provider.DoTServer, ipv6 bool) net.IP {
	if ipv6 {
		if ip := p.IP(p.counterDoTIPv6, server.IPv6); ip != nil {
			return ip
		}
		// if there is no IPv6, fall back to an IPv4 address
		// as all provider have at least an IPv4 address.
	}
	return p.IP(p.counterDoTIPv4, server.IPv4)
}

func (p *picker) DNSServer(servers []provider.DNSServer) provider.DNSServer {
	index := 0
	if nServers := len(servers); nServers > 1 {
		index = p.randIndex(p.counterDNSServer, nServers)
	}
	return servers[index]
}

func (p *picker) DNSIP(server provider.DNSServer, ipv6 bool) net.IP {
	if ipv6 {
		if ip := p.IP(p.counterDNSIPv6, server.IPv6); ip != nil {
			return ip
		}
		// if there is no IPv6, fall back to an IPv4 address
		// as all provider have at least an IPv4 address.
	}
	return p.IP(p.counterDNSIPv4, server.IPv4)
}
