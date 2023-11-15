package update

import (
	"errors"
	"fmt"
	"net/netip"
	"regexp"

	"github.com/miekg/dns"
	"github.com/qdm12/gosettings/validate"
	"github.com/qdm12/gotree"
)

type Settings struct {
	// FqdnHostnames is a list of fully qualified domain names
	// to filter out.
	FqdnHostnames []string
	// IPs is a list of IP addresses to filter out.
	IPs []netip.Addr
	// IPPrefixes is a list of IP prefixes to filter out.
	IPPrefixes []netip.Prefix
}

func (s *Settings) SetDefaults() {}

var fqdnHostRegex = regexp.MustCompile(`^([a-zA-Z0-9]|[a-zA-Z0-9_][a-zA-Z0-9\-_]{0,61}[a-zA-Z0-9_])(\.([a-zA-Z0-9]|[a-zA-Z0-9_][a-zA-Z0-9\-_]{0,61}[a-zA-Z0-9]))*\.$`) //nolint:lll

var (
	ErrFqdnHostnameNotValid = errors.New("fqdn hostname is not valid")
)

func (s Settings) Validate() (err error) {
	err = validate.AllMatchRegex(s.FqdnHostnames, fqdnHostRegex)
	if err != nil {
		return fmt.Errorf("FQDN hostnames: %w", err)
	}

	return nil
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
