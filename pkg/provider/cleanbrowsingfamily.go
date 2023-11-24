package provider

import "net/netip"

func CleanBrowsingFamily() Provider {
	return Provider{
		Name: "Cleanbrowsing Family",
		DNS: DNSServer{
			IPv4: []netip.AddrPort{
				defaultDNSIPv4AddrPort([4]byte{185, 228, 168, 168}),
				defaultDNSIPv4AddrPort([4]byte{185, 228, 169, 168}),
			},
			IPv6: []netip.AddrPort{
				defaultDNSIPv6AddrPort([16]byte{0x2a, 0xd, 0x2a, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}),
				defaultDNSIPv6AddrPort([16]byte{0x2a, 0xd, 0x2a, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}),
			},
		},
		DoT: DoTServer{
			IPv4: []netip.AddrPort{
				defaultDoTIPv4AddrPort([4]byte{185, 228, 168, 168}),
				defaultDoTIPv4AddrPort([4]byte{185, 228, 169, 168}),
			},
			IPv6: []netip.AddrPort{
				defaultDoTIPv6AddrPort([16]byte{0x2a, 0xd, 0x2a, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}),
				defaultDoTIPv6AddrPort([16]byte{0x2a, 0xd, 0x2a, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}),
			},
			Name: "family-filter-dns.cleanbrowsing.org",
		},
		// See https://cleanbrowsing.org/guides/dnsoverhttps
		DoH: DoHServer{
			URL: "https://doh.cleanbrowsing.org/doh/family-filter/",
			IPv4: []netip.Addr{
				netip.AddrFrom4([4]byte{185, 228, 168, 10}),
				netip.AddrFrom4([4]byte{185, 228, 168, 168}),
			},
		},
	}
}
