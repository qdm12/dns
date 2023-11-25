package provider

import (
	"errors"
	"fmt"
	"net/netip"
)

type Provider struct {
	Name string    `json:"name" yaml:"name"`
	DoT  DoTServer `json:"dot" yaml:"dot"`
	DoH  DoHServer `json:"doh" yaml:"doh"`
}

type DoTServer struct {
	IPv4 []netip.AddrPort `json:"ipv4" yaml:"ipv4"`
	IPv6 []netip.AddrPort `json:"ipv6" yaml:"ipv6"`
	Name string           `json:"name" yaml:"name"` // for TLS verification
}

type DoHServer struct {
	URL  string       `json:"url" yaml:"url"`
	IPv4 []netip.Addr `json:"ipv4" yaml:"ipv4"`
	IPv6 []netip.Addr `json:"ipv6" yaml:"ipv6"`
}

var (
	ErrProviderNameNotSet = errors.New("provider name not set")
	ErrDoTIPv4NotSet      = errors.New("DoT server IPv4 addresses not set")
	ErrDoTIPNotSet        = errors.New("DoT server IPv4 and IPv6 addresses not set")
	ErrDoTNameNotSet      = errors.New("DoT server name not set")
	ErrDoTPortNotSet      = errors.New("DoT server port not set")
	ErrDoHURLNotSet       = errors.New("DoH URL not set")
	ErrDoHIPv4NotSet      = errors.New("DoH server IPv4 addresses not set")
	ErrDoHIPNotSet        = errors.New("DoH server IP addresses not set")
)

func (p Provider) ValidateForDoT(ipv6 bool) (err error) {
	switch {
	case p.Name == "":
		return fmt.Errorf("%w", ErrProviderNameNotSet)
	case !ipv6 && len(p.DoT.IPv4) == 0:
		return fmt.Errorf("%w", ErrDoTIPv4NotSet)
	case ipv6 && len(p.DoT.IPv4) == 0 && len(p.DoT.IPv6) == 0:
		return fmt.Errorf("%w", ErrDoTIPNotSet)
	case p.DoT.Name == "":
		return fmt.Errorf("%w", ErrDoTNameNotSet)
	}

	err = checkAddrPorts(p.DoT.IPv4)
	if err != nil {
		return fmt.Errorf("IPv4 addresses: %w", err)
	}

	err = checkAddrPorts(p.DoT.IPv6)
	if err != nil {
		return fmt.Errorf("IPv6 addresses: %w", err)
	}

	return nil
}

func (p Provider) ValidateForDoH(ipv6 bool) (err error) {
	switch {
	case p.Name == "":
		return fmt.Errorf("%w", ErrProviderNameNotSet)
	case p.DoH.URL == "":
		return fmt.Errorf("%w", ErrDoHURLNotSet)
	case !ipv6 && len(p.DoT.IPv4) == 0:
		return fmt.Errorf("%w", ErrDoHIPv4NotSet)
	case ipv6 && len(p.DoT.IPv4) == 0 && len(p.DoT.IPv6) == 0:
		return fmt.Errorf("%w", ErrDoHIPNotSet)
	}

	err = checkAddresses(p.DoH.IPv4)
	if err != nil {
		return fmt.Errorf("IPv4 addresses: %w", err)
	}

	err = checkAddresses(p.DoH.IPv6)
	if err != nil {
		return fmt.Errorf("IPv6 addresses: %w", err)
	}

	return nil
}
