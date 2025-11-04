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
	return pickFromSlice(servers, p.rand)
}

func (p *Picker) DoTServer(servers []provider.DoTServer) provider.DoTServer {
	return pickFromSlice(servers, p.rand)
}

func (p *Picker) PlainServer(servers []provider.PlainServer) provider.PlainServer {
	return pickFromSlice(servers, p.rand)
}

//nolint:ireturn
func pickFromSlice[T any](slice []T, randSource *rand.Rand) (element T) {
	switch len(slice) {
	case 0:
		panic("slice to randomly pick from is empty")
	case 1:
		return slice[0]
	default:
		return slice[randSource.Intn(len(slice))]
	}
}

// DoTAddrPort returns a randomly picked IP address and port
// from the given DoT server. If ipv6 is true, IPv6 addresses
// are added to the pool of IP addresses to pick from, on top
// of all IPv4 addresses.
// Note IPv4 addresses are always in the pool of addresses,
// because some providers only have IPv4 addresses, and IPv4
// usually works on an IPv6 network, which is not true the other
// way around.
func (p *Picker) DoTAddrPort(server provider.DoTServer, ipv6 bool) netip.AddrPort {
	return pickFromIPs(server.IPv4, server.IPv6, ipv6, p.rand)
}

// PlainAddrPort returns a randomly picked IP address and port
// from the given plain server. If ipv6 is true, IPv6 addresses
// are added to the pool of IP addresses to pick from, on top
// of all IPv4 addresses.
// Note IPv4 addresses are always in the pool of addresses,
// because some providers only have IPv4 addresses, and IPv4
// usually works on an IPv6 network, which is not true the other
// way around.
func (p *Picker) PlainAddrPort(server provider.PlainServer, ipv6 bool) netip.AddrPort {
	return pickFromIPs(server.IPv4, server.IPv6, ipv6, p.rand)
}

func pickFromIPs(ipv4AddrPort, ipv6AddrPort []netip.AddrPort,
	ipv6 bool, rand *rand.Rand,
) netip.AddrPort {
	count := len(ipv4AddrPort)
	if ipv6 {
		count += len(ipv6AddrPort)
	}
	index := rand.Intn(count)
	if index < len(ipv4AddrPort) {
		return ipv4AddrPort[index]
	}
	return ipv6AddrPort[index-len(ipv4AddrPort)]
}
