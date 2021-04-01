package provider

import (
	"net"
	"net/url"
)

type libreDNS struct{}

func LibreDNS() Provider {
	return &libreDNS{}
}

func (l *libreDNS) String() string {
	return "LibreDNS"
}

func (l *libreDNS) DNS() DNSServer {
	// see https://libreops.cc/radicaldns.html
	return DNSServer{
		IPv4: []net.IP{{88, 198, 92, 222}},
		IPv6: []net.IP{
			{0x2a, 0x1, 0x4, 0xf8, 0x1c, 0xc, 0x82, 0xc0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1},
		},
	}
}

func (l *libreDNS) DoT() DoTServer {
	// see https://libredns.gr/
	return DoTServer{
		IPv4: []net.IP{{116, 202, 176, 26}},
		IPv6: []net.IP{},
		Name: "dot.libredns.gr",
		Port: defaultDoTPort,
	}
}

func (l *libreDNS) DoH() DoHServer {
	// See https://libredns.gr/
	return DoHServer{
		URL: &url.URL{
			Scheme: "https",
			Host:   "doh.libredns.gr",
			Path:   "/dns-query",
		},
	}
}
