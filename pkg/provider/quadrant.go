package provider

import (
	"net"
	"net/url"
)

func Quadrant() Provider {
	return Provider{
		Name: "Quadrant",
		DNS: DNSServer{
			IPv4: []net.IP{{12, 159, 2, 159}},
			IPv6: []net.IP{
				{0x20, 0x1, 0x18, 0x90, 0x14, 0xc, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x59},
			},
		},
		DoT: DoTServer{
			IPv4: []net.IP{{12, 159, 2, 159}},
			IPv6: []net.IP{
				{0x20, 0x1, 0x18, 0x90, 0x14, 0xc, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x59},
			},
			Name: "dns-tls.qis.io",
			Port: defaultDoTPort,
		},
		// See https://quadrantsec.com/quadrants_public_dns_resolver_with_tls_https_support/
		DoH: DoHServer{
			URL: url.URL{
				Scheme: "https",
				Host:   "doh.qis.io",
				Path:   "/dns-query",
			},
		},
	}
}
