package provider

import "net/netip"

func Quadrant() Provider {
	return Provider{
		Name: "Quadrant",
		DNS: DNSServer{
			IPv4: []netip.Addr{
				netip.AddrFrom4([4]byte{12, 159, 2, 159}),
			},
			IPv6: []netip.Addr{
				netip.AddrFrom16([16]byte{0x20, 0x1, 0x18, 0x90, 0x14, 0xc, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x59}),
			},
		},
		DoT: DoTServer{
			IPv4: []netip.Addr{
				netip.AddrFrom4([4]byte{12, 159, 2, 159}),
			},
			IPv6: []netip.Addr{
				netip.AddrFrom16([16]byte{0x20, 0x1, 0x18, 0x90, 0x14, 0xc, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x59}),
			},
			Name: "dns-tls.qis.io",
			Port: defaultDoTPort,
		},
		// See https://quadrantsec.com/quadrants_public_dns_resolver_with_tls_https_support/
		DoH: DoHServer{
			URL: "https://doh.qis.io/dns-query",
		},
	}
}
