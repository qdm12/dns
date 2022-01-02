package picker

import (
	"math/rand"
	"net"

	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/qdm12/golibs/crypto/random/sources/maphash"
)

type Picker struct {
	rand *rand.Rand
}

// New returns a new fast thread safe random picker
// to use for DNS servers and their IP addresses.
func New() *Picker {
	source := maphash.New()
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

func (p *Picker) DoTIP(server provider.DoTServer, ipv6 bool) net.IP {
	if ipv6 {
		if ip := p.IP(server.IPv6); ip != nil {
			return ip
		}
		// if there is no IPv6, fall back to an IPv4 address
		// as all provider have at least an IPv4 address.
	}
	return p.IP(server.IPv4)
}

func (p *Picker) DNSIP(server provider.DNSServer, ipv6 bool) net.IP {
	if ipv6 {
		if ip := p.IP(server.IPv6); ip != nil {
			return ip
		}
		// if there is no IPv6, fall back to an IPv4 address
		// as all provider have at least an IPv4 address.
	}
	return p.IP(server.IPv4)
}

func (p *Picker) IP(ips []net.IP) net.IP {
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
