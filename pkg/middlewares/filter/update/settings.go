package update

import (
	"errors"
	"fmt"
	"net/netip"
	"regexp"
	"strings"

	"github.com/miekg/dns"
	"github.com/qdm12/gosettings/validate"
	"github.com/qdm12/gotree"
)

type Settings struct {
	// FqdnHostnames is a list of fully qualified domain names
	// to filter out. If it is nil, the existing filter hostnames
	// are not updated. If it is empty, all filter hostnames are removed.
	FqdnHostnames []string
	// IPs is a list of IP addresses to filter out.
	// If it is nil, the existing filter IPs are not updated. If it is empty,
	// all filter IPs are removed.
	IPs []netip.Addr
	// IPPrefixes is a list of IP prefixes to filter out.
	// If it is nil, the existing filter IP prefixes are not updated. If it is empty,
	// all filter IP prefixes are removed.
	IPPrefixes []netip.Prefix
	// FqdnExemptFromRebindingProtection is a list of
	// fully qualified domain names that are exempt from rebinding protection.
	// If it is nil, the existing hostnames exempt from rebinding protection
	// are not updated. If it is empty, all hostnames exempt from rebinding
	// protection are removed.
	FqdnExemptFromRebindingProtection []string
	// ParentsExemptFromRebindingProtection is a list of fully qualified
	// domain names for which all their subdomains are exempt from
	// rebinding protection.
	// If it is nil, the existing parent hostnames exempt from rebinding protection
	// are not updated. If it is empty, all parent hostnames exempt from rebinding
	// protection are removed.
	ParentsExemptFromRebindingProtection []string
	// PublicFQDNsAsLocal is a list of fully qualified domain names
	// that are considered local and not public. This is useful for
	// internal services that have a public FQDN but are only accessible
	// internally. If it is nil, the existing public FQDNs considered local
	// are not updated. If it is empty, all public FQDNs considered local
	// are removed.
	PublicFQDNsAsLocal []string
}

func (s *Settings) SetDefaults() {}

var fqdnHostRegex = regexp.MustCompile(`^([a-zA-Z0-9]|[a-zA-Z0-9_][a-zA-Z0-9\-_]{0,61}[a-zA-Z0-9_])(\.([a-zA-Z0-9]|[a-zA-Z0-9_][a-zA-Z0-9\-_]{0,61}[a-zA-Z0-9]))*\.$`) //nolint:lll

var ErrFqdnHostnameNotValid = errors.New("fqdn hostname is not valid")

func (s Settings) Validate() (err error) {
	err = validate.AllMatchRegex(s.FqdnHostnames, fqdnHostRegex)
	if err != nil {
		return fmt.Errorf("FQDN hostnames: %w", err)
	}

	err = validate.AllMatchRegex(s.FqdnExemptFromRebindingProtection, fqdnHostRegex)
	if err != nil {
		return fmt.Errorf("FQDNs exempt from rebinding protection: %w", err)
	}

	err = validate.AllMatchRegex(s.ParentsExemptFromRebindingProtection, fqdnHostRegex)
	if err != nil {
		return fmt.Errorf("parent FQDNs exempt from rebinding protection: %w", err)
	}

	err = validate.AllMatchRegex(s.PublicFQDNsAsLocal, fqdnHostRegex)
	if err != nil {
		return fmt.Errorf("public FQDNs considered local: %w", err)
	}

	return nil
}

// BlockHostnames transforms the slice of hostnames given to
// FQDN hostnames and sets these to the settings.
// If the slice is nil, it sets [Settings.FqdnHostnames] to nil.
func (s *Settings) BlockHostnames(hostnames []string) {
	if hostnames == nil {
		s.FqdnHostnames = nil
		return
	}
	s.FqdnHostnames = make([]string, len(hostnames))
	for i := range hostnames {
		s.FqdnHostnames[i] = dns.Fqdn(hostnames[i])
	}
}

// SetRebindingProtectionExempt transforms the slice of hostnames given to
// FQDNs and sets these to the settings. Parent domains can be exempt by
// specifying the "*." prefix to the hostname, for example "*.example.com"
// will exempt all subdomains of example.com from rebinding protection.
// Note the wildcard cannot be used anywhere else otherwise.
// If the slice is nil, it sets [Settings.FqdnExemptFromRebindingProtection]
// and [Settings.ParentsExemptFromRebindingProtection] to nil.
func (s *Settings) SetRebindingProtectionExempt(hostnames []string) {
	if hostnames == nil {
		s.FqdnExemptFromRebindingProtection = nil
		s.ParentsExemptFromRebindingProtection = nil
		return
	}
	s.FqdnExemptFromRebindingProtection = make([]string, 0, len(hostnames))
	for _, hostname := range hostnames {
		if strings.HasPrefix(hostname, "*.") {
			parent := hostname[2:]
			s.ParentsExemptFromRebindingProtection = append(s.ParentsExemptFromRebindingProtection, dns.Fqdn(parent))
		} else {
			s.FqdnExemptFromRebindingProtection = append(s.FqdnExemptFromRebindingProtection, dns.Fqdn(hostname))
		}
	}
}

// SetPublicFQDNsAsLocal transforms the slice of hostnames given to
// FQDNs and sets these to the settings.
// If the slice is nil, it sets [Settings.PublicFQDNsAsLocal] to nil.
func (s *Settings) SetPublicFQDNsAsLocal(hostnames []string) {
	if hostnames == nil {
		s.PublicFQDNsAsLocal = nil
		return
	}
	s.PublicFQDNsAsLocal = make([]string, len(hostnames))
	for i := range hostnames {
		s.PublicFQDNsAsLocal[i] = dns.Fqdn(hostnames[i])
	}
}

func (s *Settings) String() string {
	return s.ToLinesNode().String()
}

func (s *Settings) ToLinesNode() (node *gotree.Node) { //nolint:cyclop
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

	if len(s.FqdnExemptFromRebindingProtection) > 0 {
		subNode := node.Appendf("Hostnames exempt from rebinding protection:")
		for _, fqdn := range s.FqdnExemptFromRebindingProtection {
			subNode.Appendf("%s", fqdn)
		}
	}

	if len(s.ParentsExemptFromRebindingProtection) > 0 {
		subNode := node.Appendf("Parent domains exempt from rebinding protection:")
		for _, fqdn := range s.ParentsExemptFromRebindingProtection {
			subNode.Appendf("%s", fqdn)
		}
	}

	if len(s.PublicFQDNsAsLocal) > 0 {
		subNode := node.Appendf("Public FQDNs considered local:")
		for _, fqdn := range s.PublicFQDNsAsLocal {
			subNode.Appendf("%s", fqdn)
		}
	}

	return node
}
