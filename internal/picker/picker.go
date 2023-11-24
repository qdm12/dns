package picker

import (
	"fmt"
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

func (p *Picker) DoTAddrPort(server provider.DoTServer, ipv6 bool) netip.AddrPort {
	ipVersion := "ipv4"
	serverIPs := server.IPv4
	if ipv6 {
		ipVersion = "ipv6"
		serverIPs = server.IPv6
	}

	addrPort := p.addrPort(serverIPs)
	if addrPort.IsValid() {
		return addrPort
	}
	// this should never happen since every servers
	// should have at least one IP address matching the
	// IP version given. This is more of a programming
	// safety.
	panic(fmt.Sprintf("no valid " + ipVersion + " address found"))
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
