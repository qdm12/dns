package dot

import (
	"net/netip"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/pkg/filter/update"
	"github.com/qdm12/dns/v2/pkg/provider"
)

type Middleware interface {
	Wrap(next dns.Handler) dns.Handler
}

type Filter interface {
	FilterRequest(request *dns.Msg) (blocked bool)
	FilterResponse(response *dns.Msg) (blocked bool)
	Update(settings update.Settings)
}

type Cache interface {
	Add(request, response *dns.Msg)
	Get(request *dns.Msg) (response *dns.Msg)
	Remove(request *dns.Msg)
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
	IP(ips []netip.Addr) netip.Addr
	DNSServer(servers []provider.DNSServer) provider.DNSServer
	DNSIP(server provider.DNSServer, ipv6 bool) netip.Addr
	DoTServer(servers []provider.DoTServer) provider.DoTServer
	DoTIP(server provider.DoTServer, ipv6 bool) netip.Addr
}
