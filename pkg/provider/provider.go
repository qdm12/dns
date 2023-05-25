package provider

import (
	"net"
)

const defaultDoTPort uint16 = 853

type Provider struct {
	Name string
	DNS  DNSServer
	DoT  DoTServer
	DoH  DoHServer
}

type DNSServer struct {
	IPv4 []net.IP
	IPv6 []net.IP
}

type DoTServer struct {
	IPv4 []net.IP
	IPv6 []net.IP
	Name string // for TLS verification
	Port uint16
}

type DoHServer struct {
	URL string
}
