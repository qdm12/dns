package provider

import "net/netip"

const defaultDoTPort uint16 = 853

type Provider struct {
	Name string
	DNS  DNSServer
	DoT  DoTServer
	DoH  DoHServer
}

type DNSServer struct {
	IPv4 []netip.Addr
	IPv6 []netip.Addr
}

type DoTServer struct {
	IPv4 []netip.Addr
	IPv6 []netip.Addr
	Name string // for TLS verification
	Port uint16
}

type DoHServer struct {
	URL string
}
