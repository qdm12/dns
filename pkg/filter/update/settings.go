package update

import (
	"github.com/miekg/dns"
	"github.com/qdm12/gotree"
	"inet.af/netaddr"
)

type Settings struct {
	FqdnHostnames []string
	IPs           []netaddr.IP
	IPPrefixes    []netaddr.IPPrefix
}

// BlockHostnames transforms the slice of hostnames given to
// FQDN hostnames and sets these to the settings.
func (s *Settings) BlockHostnames(hostnames []string) {
	s.FqdnHostnames = make([]string, len(hostnames))
	for i := range hostnames {
		s.FqdnHostnames[i] = dns.Fqdn(hostnames[i])
	}
}

func (s *Settings) String() string {
	return s.ToLinesNode().String()
}

func (s *Settings) ToLinesNode() (node *gotree.Node) {
	if len(s.IPs) == 0 && len(s.FqdnHostnames) == 0 &&
		len(s.IPPrefixes) == 0 {
		return gotree.New("Filter update: disabled")
	}

	node = gotree.New("Filter update settings:")

	if len(s.IPs) > 0 {
		node.Appendf("IP addresses blocked: %d", len(s.IPs))
	}

	if len(s.IPPrefixes) > 0 {
		node.Appendf("IP networks blocked: %d", len(s.IPPrefixes))
	}

	if len(s.FqdnHostnames) > 0 {
		node.Appendf("Hostnames blocked: %d", len(s.FqdnHostnames))
	}

	return node
}
