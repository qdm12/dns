package provider

import "net/netip"

const defaultDoTPort uint16 = 853

type Provider struct {
	Name string    `json:"name"`
	DNS  DNSServer `json:"dns"`
	DoT  DoTServer `json:"dot"`
	DoH  DoHServer `json:"doh"`
}

type DNSServer struct {
	IPv4 []netip.Addr `json:"ipv4"`
	IPv6 []netip.Addr `json:"ipv6"`
}

type DoTServer struct {
	IPv4 []netip.Addr `json:"ipv4"`
	IPv6 []netip.Addr `json:"ipv6"`
	Name string       `json:"name"` // for TLS verification
	Port uint16       `json:"port"`
}

type DoHServer struct {
	URL string `json:"url"`
}
