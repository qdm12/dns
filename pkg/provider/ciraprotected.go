package provider

import (
	"net"
	"net/url"
)

type ciraProtected struct{}

func CiraProtected() Provider {
	return &ciraProtected{}
}

func (c *ciraProtected) String() string {
	return "CIRA Protected"
}

func (c *ciraProtected) DNS() DNSServer {
	return DNSServer{
		IPv4: []net.IP{{149, 112, 121, 20}, {149, 112, 122, 20}},
		IPv6: []net.IP{
			{0x26, 0x20, 0x1, 0xa, 0x80, 0xbb, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x20},
			{0x26, 0x20, 0x1, 0xa, 0x80, 0xbc, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x20},
		},
	}
}

func (c *ciraProtected) DoT() DoTServer {
	return DoTServer{
		IPv4: []net.IP{{149, 112, 121, 20}, {149, 112, 122, 20}},
		IPv6: []net.IP{
			{0x26, 0x20, 0x1, 0xa, 0x80, 0xbb, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x20},
			{0x26, 0x20, 0x1, 0xa, 0x80, 0xbc, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x20},
		},
		Name: "protected.canadianshield.cira.ca",
		Port: defaultDoTPort,
	}
}

func (c *ciraProtected) DoH() DoHServer {
	return DoHServer{
		URL: &url.URL{
			Scheme: "https",
			Host:   "protected.canadianshield.cira.ca",
			Path:   "/dns-query",
		},
	}
}
