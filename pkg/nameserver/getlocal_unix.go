//go:build !js && !windows

package nameserver

import (
	"net/netip"
	"os"
	"strings"
)

func GetDNSServers() (nameservers []netip.AddrPort) {
	const filename = "/etc/resolv.conf"
	return getLocalNameservers(filename)
}

func getLocalNameservers(filename string) (nameservers []netip.AddrPort) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" || line[0] == '#' {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] != "nameserver" {
			continue
		}
		for _, field := range fields[1:] {
			ip, err := netip.ParseAddr(field)
			if err != nil {
				continue
			}

			const defaultNameserverPort = 53
			nameservers = append(nameservers,
				netip.AddrPortFrom(ip, defaultNameserverPort))
		}
	}

	return nameservers
}
