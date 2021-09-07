package blockbuilder

import (
	"inet.af/netaddr"
)

func convertIPsToString(ips []netaddr.IP) (ipStrings []string) {
	ipStrings = make([]string, len(ips))
	for i := range ips {
		ipStrings[i] = ips[i].String()
	}
	return ipStrings
}

func convertIPPrefixesToString(ipPrefixes []netaddr.IPPrefix) (ipPrefixStrings []string) {
	ipPrefixStrings = make([]string, len(ipPrefixes))
	for i := range ipPrefixes {
		ipPrefixStrings[i] = ipPrefixes[i].String()
	}
	return ipPrefixStrings
}

func convertErrorsToString(errors []error) (errorStrings []string) {
	errorStrings = make([]string, len(errors))
	for i := range errors {
		errorStrings[i] = errors[i].Error()
	}
	return errorStrings
}
