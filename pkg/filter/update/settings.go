package update

import (
	"strconv"
	"strings"

	"github.com/miekg/dns"
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
	const (
		subSection = " |--"
		indent     = "    " // used if lines already contain the subSection
	)
	return strings.Join(s.Lines(indent, subSection), "\n")
}

func (s *Settings) Lines(indent, subSection string) (lines []string) {
	if len(s.IPs) == 0 && len(s.FqdnHostnames) == 0 &&
		len(s.IPPrefixes) == 0 {
		return []string{subSection + "Filtering is disabled"}
	}

	if len(s.IPs) > 0 {
		lines = append(lines, subSection+"IP addresses blocked: "+
			strconv.Itoa(len(s.IPs)))
	}

	if len(s.IPPrefixes) > 0 {
		lines = append(lines, subSection+"IP networks blocked: "+
			strconv.Itoa(len(s.IPPrefixes)))
	}

	if len(s.FqdnHostnames) > 0 {
		lines = append(lines, subSection+"Hostnames blocked: "+
			strconv.Itoa(len(s.FqdnHostnames)))
	}

	return lines
}
