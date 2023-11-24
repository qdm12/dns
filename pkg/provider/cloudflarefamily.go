package provider

import "net/netip"

func CloudflareFamily() Provider {
	return Provider{
		Name: "Cloudflare Family",
		DNS: DNSServer{
			IPv4: []netip.AddrPort{
				defaultDNSIPv4AddrPort([4]byte{1, 1, 1, 3}),
				defaultDNSIPv4AddrPort([4]byte{1, 0, 0, 3}),
			},
			IPv6: []netip.AddrPort{
				defaultDNSIPv6AddrPort([16]byte{0x26, 0x6, 0x47, 0x0, 0x47, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x11, 0x13}),
				defaultDNSIPv6AddrPort([16]byte{0x26, 0x6, 0x47, 0x0, 0x47, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x10, 0x03}),
			},
		},
		DoT: DoTServer{
			IPv4: []netip.AddrPort{
				defaultDoTIPv4AddrPort([4]byte{1, 1, 1, 3}),
				defaultDoTIPv4AddrPort([4]byte{1, 0, 0, 3}),
			},
			IPv6: []netip.AddrPort{
				defaultDoTIPv6AddrPort([16]byte{0x26, 0x6, 0x47, 0x0, 0x47, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x11, 0x13}),
				defaultDoTIPv6AddrPort([16]byte{0x26, 0x6, 0x47, 0x0, 0x47, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x10, 0x03}),
			},
			Name: "family.cloudflare-dns.com",
		},
		// see https://developers.cloudflare.com/1.1.1.1/1.1.1.1-for-families/setup-instructions/dns-over-https
		DoH: DoHServer{
			URL: "https://family.cloudflare-dns.com/dns-query",
			IPv4: []netip.Addr{
				netip.AddrFrom4([4]byte{1, 1, 1, 3}),
				netip.AddrFrom4([4]byte{1, 0, 0, 3}),
			},
			IPv6: []netip.Addr{
				netip.AddrFrom16([16]byte{0x26, 0x6, 0x47, 0x0, 0x47, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x11, 0x13}),
				netip.AddrFrom16([16]byte{0x26, 0x6, 0x47, 0x0, 0x47, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x10, 0x03}),
			},
		},
	}
}
