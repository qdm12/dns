package blacklist

import (
	"net"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

type Settings struct {
	FqdnHostnames []string
	IPs           []net.IP
	IPNets        []*net.IPNet
}

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
	if len(s.IPs) == 0 && len(s.FqdnHostnames) == 0 {
		return []string{subSection + "Blacklisting is disabled"}
	}

	if len(s.IPs) > 0 {
		lines = append(lines, subSection+"IP addresses blocked: "+
			strconv.Itoa(len(s.IPs)))
	}

	if len(s.IPNets) > 0 {
		lines = append(lines, subSection+"IP networks blocked: "+
			strconv.Itoa(len(s.IPNets)))
	}

	if len(s.FqdnHostnames) > 0 {
		lines = append(lines, subSection+"Hostnames blocked: "+
			strconv.Itoa(len(s.FqdnHostnames)))
	}

	return lines
}
