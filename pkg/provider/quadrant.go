package provider

import "net/netip"

func Quadrant() Provider {
	return Provider{
		Name: "Quadrant",
		DNS: DNSServer{
			IPv4: []netip.AddrPort{
				defaultDNSIPv4AddrPort([4]byte{12, 159, 2, 159}),
			},
			IPv6: []netip.AddrPort{},
		},
		// See https://quadrantsec.com/blog/quadrants-public-dns-resolver-tls-https-support
		DoT: DoTServer{
			IPv4: []netip.AddrPort{
				defaultDoTIPv4AddrPort([4]byte{12, 159, 2, 159}),
			},
			IPv6: []netip.AddrPort{},
			Name: "dns-tls.qis.io",
		},
		// See https://quadrantsec.com/blog/quadrants-public-dns-resolver-tls-https-support
		DoH: DoHServer{
			URL: "https://doh.qis.io/dns-query",
		},
	}
}
