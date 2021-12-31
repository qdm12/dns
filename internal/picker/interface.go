package picker

import (
	"net"

	"github.com/qdm12/dns/pkg/provider"
)

type Interface interface {
	IP(ips []net.IP) net.IP
	DNSServer(servers []provider.DNSServer) provider.DNSServer
	DNSIP(server provider.DNSServer, ipv6 bool) net.IP
	DoT
}

type DoT interface {
	DoTServer(servers []provider.DoTServer) provider.DoTServer
	DoTIP(server provider.DoTServer, ipv6 bool) net.IP
}

type DNS interface {
	DNSServer(servers []provider.DNSServer) provider.DNSServer
	DNSIP(server provider.DNSServer, ipv6 bool) net.IP
}
