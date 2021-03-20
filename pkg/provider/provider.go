package provider

import (
	"net"
	"net/url"
)

type Provider interface {
	DNS() DNSServer
	DoT() DoTServer
	DoH() DoHServer
}

type DNSServer struct {
	IPv4 []net.IP
	IPv6 []net.IP
}

type DoTServer struct {
	IPv4 []net.IP
	IPv6 []net.IP
	Name string // for TLS verification
}

type DoHServer struct {
	URL *url.URL
}
