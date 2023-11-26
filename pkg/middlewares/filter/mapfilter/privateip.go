package mapfilter

import "net/netip"

func getPrivateIPPrefixes() (privateIPPrefixes []netip.Prefix) {
	privateCIDRs := []string{
		// IPv4 private addresses
		"127.0.0.1/8",
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16",
		// IPv6 private addresses
		"::1/128",
		"fc00::/7",
		"fe80::/10",
		// Private IPv4 addresses wrapped in IPv6
		"::ffff:7f00:1/104", // 127.0.0.1/8
		"::ffff:a00:0/104",  // 10.0.0.0/8
		"::ffff:ac10:0/108", // 172.16.0.0/12
		"::ffff:c0a8:0/112", // 192.168.0.0/16
		"::ffff:a9fe:0/112", // 169.254.0.0/16
	}
	privateIPPrefixes = make([]netip.Prefix, len(privateCIDRs))
	var err error
	for i := range privateCIDRs {
		privateIPPrefixes[i], err = netip.ParsePrefix(privateCIDRs[i])
		if err != nil {
			panic(err)
		}
	}

	return privateIPPrefixes
}
