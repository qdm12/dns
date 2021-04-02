package dot

import (
	"net"

	"github.com/qdm12/dns/pkg/provider"
	"github.com/qdm12/golibs/crypto/random/hashmap"
)

type picker struct {
	rand hashmap.Rand
}

func newPicker() *picker {
	return &picker{
		rand: hashmap.New(),
	}
}

func (p *picker) DoTServer(servers []provider.DoTServer) provider.DoTServer {
	index := 0
	if nServers := len(servers); nServers > 1 {
		index = p.rand.Intn(nServers)
	}
	return servers[index]
}

func (p *picker) IP(ips []net.IP) net.IP {
	switch len(ips) {
	case 0:
		return nil
	case 1:
		return ips[0]
	default:
		index := p.rand.Intn(len(ips))
		return ips[index]
	}
}

func (p *picker) DoTIP(server provider.DoTServer, ipv6 bool) net.IP {
	if ipv6 {
		if ip := p.IP(server.IPv6); ip != nil {
			return ip
		}
		// if there is no IPv6, fall back to an IPv4 address
		// as all provider have at least an IPv4 address.
	}
	return p.IP(server.IPv4)
}

func (p *picker) DNSServer(servers []provider.DNSServer) provider.DNSServer {
	index := 0
	if nServers := len(servers); nServers > 1 {
		index = p.rand.Intn(nServers)
	}
	return servers[index]
}

func (p *picker) DNSIP(server provider.DNSServer, ipv6 bool) net.IP {
	if ipv6 {
		if ip := p.IP(server.IPv6); ip != nil {
			return ip
		}
		// if there is no IPv6, fall back to an IPv4 address
		// as all provider have at least an IPv4 address.
	}
	return p.IP(server.IPv4)
}
