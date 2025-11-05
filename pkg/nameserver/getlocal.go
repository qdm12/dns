package nameserver

import "net/netip"

func GetPrivateDNSServers() (nameservers []netip.AddrPort) {
	allNameservers := GetDNSServers()
	nameservers = make([]netip.AddrPort, 0, len(allNameservers))
	for _, server := range allNameservers {
		if server.Addr().IsPrivate() || server.Addr().IsLoopback() {
			nameservers = append(nameservers, server)
		}
	}
	return nameservers
}
