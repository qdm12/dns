package nameserver

import "net/netip"

func GetPrivateDNSServers() (nameservers []netip.Addr, err error) {
	return GetPrivateDNSServersWithPublicCIDRsAsLocal(nil)
}

func GetPrivateDNSServersWithPublicCIDRsAsLocal(publicCIDRsAsLocal []netip.Prefix) (
	nameservers []netip.Addr, err error,
) {
	allNameservers, err := GetDNSServers()
	if err != nil {
		return nil, err
	}
	nameservers = make([]netip.Addr, 0, len(allNameservers))
	for _, server := range allNameservers {
		if server.IsPrivate() || server.IsLoopback() ||
			addrContainedByAnyPrefix(server, publicCIDRsAsLocal) {
			nameservers = append(nameservers, server)
		}
	}
	return nameservers, nil
}

func addrContainedByAnyPrefix(address netip.Addr, prefixes []netip.Prefix) bool {
	for _, prefix := range prefixes {
		if prefix.Contains(address) {
			return true
		}
	}

	return false
}
