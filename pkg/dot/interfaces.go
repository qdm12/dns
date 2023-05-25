package dot

import (
	"net/netip"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/pkg/provider"
)

type Middleware interface {
	Wrap(next dns.Handler) dns.Handler
}

type Picker interface {
	IP(ips []netip.Addr) netip.Addr
	DNSServer(servers []provider.DNSServer) provider.DNSServer
	DNSIP(server provider.DNSServer, ipv6 bool) netip.Addr
	DoTServer(servers []provider.DoTServer) provider.DoTServer
	DoTIP(server provider.DoTServer, ipv6 bool) netip.Addr
}
