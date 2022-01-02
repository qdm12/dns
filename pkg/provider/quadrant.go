package provider

import (
	"net"
	"net/url"
)

type quadrant struct{}

func Quadrant() Provider {
	return &quadrant{}
}

func (q *quadrant) String() string {
	return "Quadrant"
}

func (q *quadrant) DNS() DNSServer {
	return DNSServer{
		IPv4: []net.IP{{12, 159, 2, 159}},
		IPv6: []net.IP{
			{0x20, 0x1, 0x18, 0x90, 0x14, 0xc, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x59},
		},
	}
}

func (q *quadrant) DoT() DoTServer {
	return DoTServer{
		IPv4: []net.IP{{12, 159, 2, 159}},
		IPv6: []net.IP{
			{0x20, 0x1, 0x18, 0x90, 0x14, 0xc, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x59},
		},
		Name: "dns-tls.qis.io",
		Port: defaultDoTPort,
	}
}

func (q *quadrant) DoH() DoHServer {
	// See https://quadrantsec.com/quadrants_public_dns_resolver_with_tls_https_support/
	return DoHServer{
		URL: &url.URL{
			Scheme: "https",
			Host:   "doh.qis.io",
			Path:   "/dns-query",
		},
	}
}
