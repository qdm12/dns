package filter

import (
	"sort"
	"strconv"
	"strings"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/pkg/filter/metrics"
	"github.com/qdm12/dns/pkg/filter/metrics/noop"
	"inet.af/netaddr"
)

type Settings struct {
	FqdnHostnames []string
	IPs           []netaddr.IP
	IPPrefixes    []netaddr.IPPrefix
	Metrics       metrics.Interface
}

func (s *Settings) setDefaults() {
	if s.Metrics == nil {
		s.Metrics = noop.New()
	}
}

// BlockHostnames transforms the slice of hostnames given to
// FQDN hostnames and sets these to the settings.
func (s *Settings) BlockHostnames(hostnames []string) {
	s.FqdnHostnames = make([]string, len(hostnames))
	for i := range hostnames {
		s.FqdnHostnames[i] = dns.Fqdn(hostnames[i])
	}
}

// AddBlockHostnames transforms the slice of hostnames given to
// FQDN hostnames and adds the new hostnames to the settings,
// removing any duplicate.
func (s *Settings) AddBlockHostnames(hostnames []string) {
	fqdnHostnames := make([]string, len(hostnames))
	for i := range hostnames {
		fqdnHostnames[i] = dns.Fqdn(hostnames[i])
	}
	set := make(map[string]struct{}, len(s.FqdnHostnames)+len(fqdnHostnames))
	for _, host := range s.FqdnHostnames {
		set[host] = struct{}{}
	}
	for _, host := range fqdnHostnames {
		set[host] = struct{}{}
	}
	s.FqdnHostnames = make([]string, 0, len(set))
	for host := range set {
		s.FqdnHostnames = append(s.FqdnHostnames, host)
	}
	sort.Slice(s.FqdnHostnames, func(i, j int) bool {
		return s.FqdnHostnames[i] < s.FqdnHostnames[j]
	})
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
