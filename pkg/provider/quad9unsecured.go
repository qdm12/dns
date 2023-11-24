package provider

import "net/netip"

func Quad9Unsecured() Provider {
	return Provider{
		Name: "Quad9 Unsecured",
		DoT: DoTServer{
			IPv4: []netip.AddrPort{
				defaultDoTIPv4AddrPort([4]byte{9, 9, 9, 10}),
				defaultDoTIPv4AddrPort([4]byte{149, 112, 112, 10}),
			},
			IPv6: []netip.AddrPort{
				defaultDoTIPv6AddrPort([16]byte{0x26, 0x20, 0x0, 0xfe, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x10}),
				defaultDoTIPv6AddrPort([16]byte{0x26, 0x20, 0x0, 0xfe, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x0, 0x10}),
			},
			Name: "dns10.quad9.net",
		},
		// See https://www.quad9.net/news/blog/doh-with-quad9-dns-servers/
		DoH: DoHServer{
			URL: "https://dns10.quad9.net/dns-query",
			IPv4: []netip.Addr{
				netip.AddrFrom4([4]byte{9, 9, 9, 10}),
				netip.AddrFrom4([4]byte{149, 112, 112, 10}),
			},
			IPv6: []netip.Addr{
				netip.AddrFrom16([16]byte{0x26, 0x20, 0x0, 0xfe, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x10}),
				netip.AddrFrom16([16]byte{0x26, 0x20, 0x0, 0xfe, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x0, 0x10}),
			},
		},
	}
}
