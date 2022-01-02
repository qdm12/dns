package blockbuilder

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/qdm12/dns/internal/settings/defaults"
	"github.com/qdm12/gotree"
	"inet.af/netaddr"
)

type Settings struct {
	Client               *http.Client
	BlockMalicious       *bool
	BlockAds             *bool
	BlockSurveillance    *bool
	AllowedHosts         []string
	AllowedIPs           []netaddr.IP
	AllowedIPPrefixes    []netaddr.IPPrefix
	AddBlockedHosts      []string
	AddBlockedIPs        []netaddr.IP
	AddBlockedIPPrefixes []netaddr.IPPrefix
}

func (s *Settings) SetDefaults() {
	s.Client = defaults.HTTPClient(s.Client, http.DefaultClient)
	s.BlockMalicious = defaults.BoolPtr(s.BlockMalicious, true)
	s.BlockAds = defaults.BoolPtr(s.BlockAds, false)
	s.BlockSurveillance = defaults.BoolPtr(s.BlockSurveillance, false)
}

var hostRegex = regexp.MustCompile(`^([a-zA-Z0-9]|[a-zA-Z0-9_][a-zA-Z0-9\-_]{0,61}[a-zA-Z0-9_])(\.([a-zA-Z0-9]|[a-zA-Z0-9_][a-zA-Z0-9\-_]{0,61}[a-zA-Z0-9]))*$`) //nolint:lll

var (
	ErrAllowedHostNotValid = errors.New("allowed host is not valid")
	ErrBlockedHostNotValid = errors.New("blocked host is not valid")
)

func (s Settings) Validate() (err error) {
	for _, host := range s.AllowedHosts {
		if !hostRegex.MatchString(host) {
			return fmt.Errorf("%w: %s", ErrAllowedHostNotValid, host)
		}
	}

	for _, host := range s.AddBlockedHosts {
		if !hostRegex.MatchString(host) {
			return fmt.Errorf("%w: %s", ErrBlockedHostNotValid, host)
		}
	}

	return nil
}

func (s *Settings) String() string {
	return s.ToLinesNode().String()
}

func (s *Settings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("Filter build settings:")

	var blockedCategories []string
	if *s.BlockMalicious {
		blockedCategories = append(blockedCategories, "malicious")
	}
	if *s.BlockSurveillance {
		blockedCategories = append(blockedCategories, "surveillance")
	}
	if *s.BlockAds {
		blockedCategories = append(blockedCategories, "ads")
	}

	node.Appendf("Blocked categories: %s", strings.Join(blockedCategories, ", "))

	if len(s.AllowedHosts) > 0 {
		node.Appendf("Hostnames unblocked: %d", len(s.AllowedHosts))
	}

	if len(s.AllowedIPs) > 0 {
		node.Appendf("IP addresses unblocked: %d", len(s.AllowedIPs))
	}

	if len(s.AllowedIPPrefixes) > 0 {
		node.Appendf("IP networks unblocked: %d", len(s.AllowedIPPrefixes))
	}

	if len(s.AddBlockedHosts) > 0 {
		node.Appendf("Additional hostnames blocked: %d", len(s.AddBlockedHosts))
	}

	if len(s.AddBlockedIPs) > 0 {
		node.Appendf("Additional IP addresses blocked: %d", len(s.AddBlockedIPs))
	}

	if len(s.AddBlockedIPPrefixes) > 0 {
		node.Appendf("Additional IP networks blocked: %d", len(s.AddBlockedIPPrefixes))
	}

	return node
}
