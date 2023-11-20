package dot

import (
	"net/netip"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/pkg/provider"
)

type Middleware interface {
	Wrap(next dns.Handler) dns.Handler
}

type Metrics interface {
	DoTDialInc(provider, address, outcome string)
	DNSDialInc(address, outcome string)
}

type Logger interface {
	Debug(s string)
	Info(s string)
	Warner
	Error(s string)
}

type Warner interface {
	Warn(s string)
}

type Picker interface {
	DNSServer(servers []provider.DNSServer) provider.DNSServer
	DNSAddrPort(server provider.DNSServer, ipv6 bool) netip.AddrPort
	DoTServer(servers []provider.DoTServer) provider.DoTServer
	DoTAddrPort(server provider.DoTServer, ipv6 bool) netip.AddrPort
}
