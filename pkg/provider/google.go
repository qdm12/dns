package provider

import "net/netip"

func Google() Provider {
	return Provider{
		Name: "Google",
		DNS: DNSServer{
			IPv4: []netip.AddrPort{
				defaultDNSIPv4AddrPort([4]byte{8, 8, 8, 8}),
				defaultDNSIPv4AddrPort([4]byte{8, 8, 4, 4}),
			},
			IPv6: []netip.AddrPort{
				defaultDNSIPv6AddrPort([16]byte{0x20, 0x1, 0x48, 0x60, 0x48, 0x60, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x88, 0x88}),
				defaultDNSIPv6AddrPort([16]byte{0x20, 0x1, 0x48, 0x60, 0x48, 0x60, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x88, 0x44}),
			},
		},
		DoT: DoTServer{
			IPv4: []netip.AddrPort{
				defaultDoTIPv4AddrPort([4]byte{8, 8, 8, 8}),
				defaultDoTIPv4AddrPort([4]byte{8, 8, 4, 4}),
			},
			IPv6: []netip.AddrPort{
				defaultDoTIPv6AddrPort([16]byte{0x20, 0x1, 0x48, 0x60, 0x48, 0x60, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x88, 0x88}),
				defaultDoTIPv6AddrPort([16]byte{0x20, 0x1, 0x48, 0x60, 0x48, 0x60, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x88, 0x44}),
			},
			Name: "dns.google",
		},
		// See https://developers.google.com/speed/public-dns/docs/doh
		DoH: DoHServer{
			URL: "https://dns.google/dns-query",
			IPv4: []netip.Addr{
				netip.AddrFrom4([4]byte{8, 8, 8, 8}),
				netip.AddrFrom4([4]byte{8, 8, 4, 4}),
			},
			IPv6: []netip.Addr{
				netip.AddrFrom16([16]byte{0x20, 0x1, 0x48, 0x60, 0x48, 0x60, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x88, 0x88}),
				netip.AddrFrom16([16]byte{0x20, 0x1, 0x48, 0x60, 0x48, 0x60, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x88, 0x44}),
			},
		},
	}
}
