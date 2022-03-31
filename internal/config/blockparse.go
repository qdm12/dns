package config

import (
	"errors"
	"fmt"

	"github.com/qdm12/golibs/verification"
	"inet.af/netaddr"
)

var errHostnameInvalid = errors.New("hostname is invalid")

func checkHostnames(verifier verification.Verifier, hostnames []string) (err error) {
	for _, hostname := range hostnames {
		if !verifier.MatchHostname(hostname) {
			return fmt.Errorf("%w: %s", errHostnameInvalid, hostname)
		}
	}
	return nil
}

var errIPStringInvalid = errors.New("IP address string is invalid")

func parseIPStrings(ipStrings []string) (ips []netaddr.IP, err error) {
	ips = make([]netaddr.IP, len(ipStrings))

	for i, ipString := range ipStrings {
		ips[i], err = netaddr.ParseIP(ipString)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", errIPStringInvalid, err)
		}
	}

	return ips, nil
}

var errIPPrefixStringInvalid = errors.New("IP prefix CIDR string is invalid")

func parseIPPrefixStrings(ipPrefixStrings []string) (ipPrefixes []netaddr.IPPrefix, err error) {
	ipPrefixes = make([]netaddr.IPPrefix, len(ipPrefixStrings))

	for i, ipPrefixString := range ipPrefixStrings {
		ipPrefixes[i], err = netaddr.ParseIPPrefix(ipPrefixString)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", errIPPrefixStringInvalid, err)
		}
	}

	return ipPrefixes, nil
}
