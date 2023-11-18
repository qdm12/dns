package provider

import (
	"errors"
	"fmt"
	"net/netip"
)

const defaultDoTPort uint16 = 853

type Provider struct {
	Name string    `json:"name" yaml:"name"`
	DNS  DNSServer `json:"dns" yaml:"dns"`
	DoT  DoTServer `json:"dot" yaml:"dot"`
	DoH  DoHServer `json:"doh" yaml:"doh"`
}

type DNSServer struct {
	IPv4 []netip.Addr `json:"ipv4" yaml:"ipv4"`
	IPv6 []netip.Addr `json:"ipv6" yaml:"ipv6"`
}

type DoTServer struct {
	IPv4 []netip.Addr `json:"ipv4" yaml:"ipv4"`
	IPv6 []netip.Addr `json:"ipv6" yaml:"ipv6"`
	Name string       `json:"name" yaml:"name"` // for TLS verification
	Port uint16       `json:"port" yaml:"port"`
}

type DoHServer struct {
	URL string `json:"url" yaml:"url"`
}

var (
	ErrProviderNameNotSet = errors.New("provider name not set")
	ErrDNSIPv4NotSet      = errors.New("DNS plaintext server IPv4 address not set")
	ErrDNSIPv6NotSet      = errors.New("DNS plaintext server IPv6 address not set")
	ErrDoTIPv4NotSet      = errors.New("DoT server IPv4 address not set")
	ErrDoTIPv6NotSet      = errors.New("DoT server IPv6 address not set")
	ErrDoTNameNotSet      = errors.New("DoT server name not set")
	ErrDoTPortNotSet      = errors.New("DoT server port not set")
	ErrDoHURLNotSet       = errors.New("DoH URL not set")
)

func (p Provider) ValdidateForPlaintext() (err error) {
	switch {
	case p.Name == "":
		return fmt.Errorf("%w", ErrProviderNameNotSet)
	case len(p.DNS.IPv4) == 0:
		return fmt.Errorf("%w", ErrDNSIPv4NotSet)
	case len(p.DNS.IPv6) == 0:
		return fmt.Errorf("%w", ErrDNSIPv6NotSet)
	}

	return nil
}

func (p Provider) ValidateForDoT() (err error) {
	if p.Name == "" {
		return fmt.Errorf("%w", ErrProviderNameNotSet)
	}

	switch {
	case len(p.DoT.IPv4) == 0:
		return fmt.Errorf("%w", ErrDoTIPv4NotSet)
	case len(p.DoT.IPv6) == 0:
		return fmt.Errorf("%w", ErrDoTIPv6NotSet)
	case p.DoT.Name == "":
		return fmt.Errorf("%w: %s", ErrDoTNameNotSet, p.DoT.Name)
	case p.DoT.Port == 0:
		return fmt.Errorf("%w", ErrDoTPortNotSet)
	}

	return nil
}

func (p Provider) ValidateForDoH() (err error) {
	switch {
	case p.Name == "":
		return fmt.Errorf("%w", ErrProviderNameNotSet)
	case p.DoH.URL == "":
		return fmt.Errorf("%w", ErrDoHURLNotSet)
	}

	return nil
}
