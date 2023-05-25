package provider

import (
	"net"
)

func LibreDNS() Provider {
	return Provider{
		Name: "LibreDNS",
		// see https://libreops.cc/radicaldns.html
		DNS: DNSServer{
			IPv4: []net.IP{{88, 198, 92, 222}},
			IPv6: []net.IP{
				{0x2a, 0x1, 0x4, 0xf8, 0x1c, 0xc, 0x82, 0xc0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1},
			},
		},
		// see https://libredns.gr/
		DoT: DoTServer{
			IPv4: []net.IP{{116, 202, 176, 26}},
			IPv6: []net.IP{},
			Name: "dot.libredns.gr",
			Port: defaultDoTPort,
		},
		// see https://libredns.gr/
		DoH: DoHServer{
			URL: "https://doh.libredns.gr/dns-query",
		},
	}
}
