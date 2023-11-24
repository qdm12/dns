package provider

import "net/netip"

func OpenDNS() Provider {
	return Provider{
		Name: "OpenDNS",
		DoT: DoTServer{
			IPv4: []netip.AddrPort{
				defaultDoTIPv4AddrPort([4]byte{208, 67, 222, 222}),
				defaultDoTIPv4AddrPort([4]byte{208, 67, 220, 220}),
			},
			IPv6: []netip.AddrPort{
				defaultDoTIPv6AddrPort([16]byte{0x26, 0x20, 0x1, 0x19, 0x0, 0x35, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x35}),
				defaultDoTIPv6AddrPort([16]byte{0x26, 0x20, 0x1, 0x19, 0x0, 0x53, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x53}),
			},
			Name: "dns.opendns.com",
		},
		// See https://support.opendns.com/hc/en-us/articles/360038086532-Using-DNS-over-HTTPS-DoH-with-OpenDNS
		DoH: DoHServer{
			URL: "https://dns.opendns.com/dns-query",
			IPv4: []netip.Addr{
				netip.AddrFrom4([4]byte{208, 67, 222, 222}),
				netip.AddrFrom4([4]byte{208, 67, 220, 220}),
			},
			IPv6: []netip.Addr{
				netip.AddrFrom16([16]byte{0x26, 0x20, 0x1, 0x19, 0x0, 0x35, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x35}),
				netip.AddrFrom16([16]byte{0x26, 0x20, 0x1, 0x19, 0x0, 0x53, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x53}),
			},
		},
	}
}
