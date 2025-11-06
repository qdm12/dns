package nameserver

import "net/netip"

func GetPrivateDNSServers() (nameservers []netip.Addr, err error) {
	allNameservers, err := GetDNSServers()
	if err != nil {
		return nil, err
	}
	nameservers = make([]netip.Addr, 0, len(allNameservers))
	for _, server := range allNameservers {
		if server.IsPrivate() || server.IsLoopback() {
			nameservers = append(nameservers, server)
		}
	}
	return nameservers, nil
}
