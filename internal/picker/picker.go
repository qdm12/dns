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

func pickFromSlice[T any](slice []T, randSource *rand.Rand) (element T) { //nolint:ireturn
	switch len(slice) {
	case 0:
		panic("slice to randomly pick from is empty")
	case 1:
		return slice[0]
	default:
		return slice[randSource.Intn(len(slice))]
	}
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
	count := len(server.IPv4)
	if ipv6 {
		count += len(server.IPv6)
	}
	index := p.rand.Intn(count)
	if index < len(server.IPv4) {
		return server.IPv4[index]
	}
	return server.IPv6[index-len(server.IPv4)]
}
