package provider

import (
	"net"
	"net/url"
)

type ciraFamily struct{}

func CiraFamily() Provider {
	return &ciraFamily{}
}

func (c *ciraFamily) String() string {
	return "CIRA Family"
}

func (c *ciraFamily) DNS() DNSServer {
	return DNSServer{
		IPv4: []net.IP{{149, 112, 121, 30}, {149, 112, 122, 30}},
		IPv6: []net.IP{
			{0x26, 0x20, 0x1, 0xa, 0x80, 0xbb, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x30},
			{0x26, 0x20, 0x1, 0xa, 0x80, 0xbc, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x30},
		},
	}
}

func (c *ciraFamily) DoT() DoTServer {
	return DoTServer{
		IPv4: []net.IP{{149, 112, 121, 30}, {149, 112, 122, 30}},
		IPv6: []net.IP{
			{0x26, 0x20, 0x1, 0xa, 0x80, 0xbb, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x30},
			{0x26, 0x20, 0x1, 0xa, 0x80, 0xbc, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x30},
		},
		Name: "family.canadianshield.cira.ca",
		Port: defaultDoTPort,
	}
}

func (c *ciraFamily) DoH() DoHServer {
	return DoHServer{
		URL: &url.URL{
			Scheme: "https",
			Host:   "family.canadianshield.cira.ca",
			Path:   "/dns-query",
		},
	}
}
