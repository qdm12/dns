package provider

import "net/netip"

func Quadrant() Provider {
	return Provider{
		Name: "Quadrant",
		// See https://quadrantsec.com/blog/quadrants-public-dns-resolver-tls-https-support
		DoT: DoTServer{
			IPv4: []netip.AddrPort{
				defaultDoTIPv4AddrPort([4]byte{12, 159, 2, 159}),
			},
			Name: "dns-tls.qis.io",
		},
		// See https://quadrantsec.com/blog/quadrants-public-dns-resolver-tls-https-support
		DoH: DoHServer{
			URL: "https://doh.qis.io/dns-query",
			IPv4: []netip.Addr{
				netip.AddrFrom4([4]byte{12, 159, 2, 159}),
			},
		},
	}
}
