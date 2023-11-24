package provider

import "net/netip"

func LibreDNS() Provider {
	return Provider{
		Name: "LibreDNS",
		// see https://libreops.cc/radicaldns.html
		// see https://libredns.gr/
		DoT: DoTServer{
			IPv4: []netip.AddrPort{
				defaultDoTIPv4AddrPort([4]byte{116, 202, 176, 26}),
			},
			IPv6: []netip.AddrPort{
				defaultDoTIPv6AddrPort([16]byte{0x2a, 0x1, 0x4, 0xf8, 0x1c, 0xc, 0x82, 0x74, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1}),
			},
			Name: "dot.libredns.gr",
		},
		// see https://libredns.gr/
		DoH: DoHServer{
			URL: "https://doh.libredns.gr/dns-query",
			IPv4: []netip.Addr{
				netip.AddrFrom4([4]byte{116, 202, 176, 26}),
			},
			IPv6: []netip.Addr{
				netip.AddrFrom16([16]byte{0x2a, 0x1, 0x4, 0xf8, 0x1c, 0xc, 0x82, 0x74, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1}),
			},
		},
	}
}
