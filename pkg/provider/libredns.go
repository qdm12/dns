package provider

import "net/netip"

func LibreDNS() Provider {
	return Provider{
		Name: "LibreDNS",
		// see https://libreops.cc/radicaldns.html
		DNS: DNSServer{
			IPv4: []netip.AddrPort{
				defaultDNSIPv4AddrPort([4]byte{88, 198, 92, 222}),
				defaultDNSIPv4AddrPort([4]byte{192, 71, 166, 92}),
			},
			IPv6: []netip.AddrPort{
				defaultDNSIPv6AddrPort([16]byte{0x2a, 0x1, 0x4, 0xf8, 0x1c, 0xc, 0x82, 0xc0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1}),
				defaultDNSIPv6AddrPort([16]byte{0x2a, 0x3, 0xf, 0x80, 0x0, 0x30, 0x1, 0x92, 0x0, 0x71, 0x1, 0x66, 0x0, 0x92, 0x0, 0x1}),
			},
		},
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
		},
	}
}
