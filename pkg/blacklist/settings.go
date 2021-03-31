package blacklist

import (
	"net"

	"github.com/miekg/dns"
)

type Settings struct {
	FqdnHostnames []string
	IPs           []net.IP
}

func (s *Settings) setDefaults() {}

func (s *Settings) BlockHostnames(hostnames []string) {
	s.FqdnHostnames = make([]string, len(hostnames))
	for i := range hostnames {
		s.FqdnHostnames[i] = dns.Fqdn(hostnames[i])
	}
}
