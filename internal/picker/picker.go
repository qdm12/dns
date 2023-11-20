package picker

import (
	"math/rand"
	"net/netip"

	"github.com/qdm12/dns/v2/pkg/provider"
)

type Picker struct {
	rand *rand.Rand
}

// New returns a new fast thread safe random picker
// to use for DNS servers and their IP addresses.
func New() *Picker {
	source := newMaphashSource()
	return &Picker{
		rand: rand.New(source), //nolint:gosec
	}
}

func (p *Picker) DoHServer(servers []provider.DoHServer) provider.DoHServer {
	index := 0
	if nServers := len(servers); nServers > 1 {
		index = p.rand.Intn(nServers)
	}
	return servers[index]
}

func (p *Picker) DoTServer(servers []provider.DoTServer) provider.DoTServer {
	index := 0
	if nServers := len(servers); nServers > 1 {
		index = p.rand.Intn(nServers)
	}
	return servers[index]
}

func (p *Picker) DNSServer(servers []provider.DNSServer) provider.DNSServer {
	index := 0
	if nServers := len(servers); nServers > 1 {
		index = p.rand.Intn(nServers)
	}
	return servers[index]
}

func (p *Picker) DoTAddrPort(server provider.DoTServer, ipv6 bool) netip.AddrPort {
	if ipv6 {
		if ip := p.addrPort(server.IPv6); ip.IsValid() {
			return ip
		}
		// if there is no IPv6, fall back to an IPv4 address
		// as all provider have at least an IPv4 address.
	}
	return p.addrPort(server.IPv4)
}

func (p *Picker) DNSAddrPort(server provider.DNSServer, ipv6 bool) netip.AddrPort {
	if ipv6 {
		if ip := p.addrPort(server.IPv6); ip.IsValid() {
			return ip
		}
		// if there is no IPv6, fall back to an IPv4 address
		// as all provider have at least an IPv4 address.
	}
	return p.addrPort(server.IPv4)
}

func (p *Picker) addrPort(addrPorts []netip.AddrPort) netip.AddrPort {
	switch len(addrPorts) {
	case 0:
		return netip.AddrPort{}
	case 1:
		return addrPorts[0]
	default:
		index := p.rand.Intn(len(addrPorts))
		return addrPorts[index]
	}
}
