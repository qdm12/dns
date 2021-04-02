package blacklist

import "net"

func convertIPsToString(ips []net.IP) (ipStrings []string) {
	ipStrings = make([]string, len(ips))
	for i := range ips {
		ipStrings[i] = ips[i].String()
	}
	return ipStrings
}

func convertIPNetsToString(ipNets []*net.IPNet) (ipNetStrings []string) {
	ipNetStrings = make([]string, len(ipNets))
	for i := range ipNets {
		ipNetStrings[i] = ipNets[i].String()
	}
	return ipNetStrings
}

func convertErrorsToString(errors []error) (errorStrings []string) {
	errorStrings = make([]string, len(errors))
	for i := range errors {
		errorStrings[i] = errors[i].Error()
	}
	return errorStrings
}
