package dot

import (
	"net"

	"github.com/qdm12/dns/v2/pkg/provider"
)

type Picker interface {
	IP(ips []net.IP) net.IP
	DNSServer(servers []provider.DNSServer) provider.DNSServer
	DNSIP(server provider.DNSServer, ipv6 bool) net.IP
	DoTServer(servers []provider.DoTServer) provider.DoTServer
	DoTIP(server provider.DoTServer, ipv6 bool) net.IP
}
