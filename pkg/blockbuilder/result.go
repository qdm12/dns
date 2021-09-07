package blockbuilder

import "inet.af/netaddr"

type Result struct {
	BlockedHostnames  []string
	BlockedIPs        []netaddr.IP
	BlockedIPPrefixes []netaddr.IPPrefix
	// Errors contains all errors encountered.
	Errors []error
}
