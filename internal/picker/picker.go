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

// DotAddrPort returns a randomly picked IP address and port
// from the given DoT server. If ipv6 is true, IPv6 addresses
// are added to the pool of IP addresses to pick from, on top
// of all IPv4 addresses.
// Note IPv4 addresses are always in the pool of addresses,
// because some providers only have IPv4 addresses, and IPv4
// usually works on an IPv6 network, which is not true the other
// way around.
func (p *Picker) DoTAddrPort(server provider.DoTServer, ipv6 bool) netip.AddrPort {
	totalSize := len(server.IPv4)
	if ipv6 {
		totalSize += len(server.IPv6)
	}
	serverIPs := make([]netip.AddrPort, 0, totalSize)
	serverIPs = append(serverIPs, server.IPv4...)
	if ipv6 {
		serverIPs = append(serverIPs, server.IPv6...)
	}

	addrPort := p.addrPort(serverIPs)
	if addrPort.IsValid() {
		return addrPort
	}
	// this should never happen since every servers
	// should have at least one IP address matching the
	// IP version given. This is more of a programming
	// safety.
	ipVersions := "IPv4"
	if ipv6 {
		ipVersions += " or IPv6"
	}
	panic("no valid " + ipVersions + " address found")
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
