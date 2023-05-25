package blockbuilder

import (
	"net/netip"
)

type Result struct {
	BlockedHostnames  []string
	BlockedIPs        []netip.Addr
	BlockedIPPrefixes []netip.Prefix
	// Errors contains all errors encountered.
	Errors []error
}
