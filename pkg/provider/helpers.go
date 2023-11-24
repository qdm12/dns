package provider

import (
	"errors"
	"fmt"
	"net/netip"
)

const (
	defaultDNSPort uint16 = 53
	defaultDoTPort uint16 = 853
)

func defaultDNSIPv4AddrPort(bytes [4]byte) netip.AddrPort {
	return netip.AddrPortFrom(netip.AddrFrom4(bytes), defaultDNSPort)
}

func defaultDNSIPv6AddrPort(bytes [16]byte) netip.AddrPort {
	return netip.AddrPortFrom(netip.AddrFrom16(bytes), defaultDNSPort)
}

func defaultDoTIPv4AddrPort(bytes [4]byte) netip.AddrPort {
	return netip.AddrPortFrom(netip.AddrFrom4(bytes), defaultDoTPort)
}

func defaultDoTIPv6AddrPort(bytes [16]byte) netip.AddrPort {
	return netip.AddrPortFrom(netip.AddrFrom16(bytes), defaultDoTPort)
}

var (
	ErrIPNotSet        = errors.New("IP address is not set")
	ErrIPIsUnspecified = errors.New("IP address is unspecified")
	ErrPortNotSet      = errors.New("port is not set")
)

func checkAddresses(addresses []netip.Addr) (err error) {
	for i, address := range addresses {
		switch {
		case !address.IsValid():
			return fmt.Errorf("address %d of %d: %w",
				i+1, len(addresses), ErrIPNotSet)
		case address.IsUnspecified():
			return fmt.Errorf("address %d of %d: %w",
				i+1, len(addresses), ErrIPIsUnspecified)
		}
	}

	return nil
}

func checkAddrPorts(addrPorts []netip.AddrPort) (err error) {
	for i, addrPort := range addrPorts {
		ip := addrPort.Addr()
		port := addrPort.Port()
		switch {
		case !ip.IsValid():
			return fmt.Errorf("address port %d of %d: %w",
				i+1, len(addrPorts), ErrIPNotSet)
		case ip.IsUnspecified():
			return fmt.Errorf("address port %d of %d: %w",
				i+1, len(addrPorts), ErrIPIsUnspecified)
		case port == 0:
			return fmt.Errorf("address port %d of %d: %w",
				i+1, len(addrPorts), ErrPortNotSet)
		}
	}

	return nil
}
