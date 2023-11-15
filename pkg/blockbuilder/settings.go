package blockbuilder

import (
	"fmt"
	"net/http"
	"net/netip"
	"regexp"
	"strings"

	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/validate"
	"github.com/qdm12/gotree"
)

type Settings struct {
	Client               *http.Client
	BlockMalicious       *bool
	BlockAds             *bool
	BlockSurveillance    *bool
	AllowedHosts         []string
	AllowedIPs           []netip.Addr
	AllowedIPPrefixes    []netip.Prefix
	AddBlockedHosts      []string
	AddBlockedIPs        []netip.Addr
	AddBlockedIPPrefixes []netip.Prefix
}

func (s *Settings) SetDefaults() {
	s.Client = gosettings.DefaultPointerRaw(s.Client, http.DefaultClient)
	s.BlockMalicious = gosettings.DefaultPointer(s.BlockMalicious, false)
	s.BlockAds = gosettings.DefaultPointer(s.BlockAds, false)
	s.BlockSurveillance = gosettings.DefaultPointer(s.BlockSurveillance, false)
}

var hostRegex = regexp.MustCompile(`^([a-zA-Z0-9]|[a-zA-Z0-9_][a-zA-Z0-9\-_]{0,61}[a-zA-Z0-9_])(\.([a-zA-Z0-9]|[a-zA-Z0-9_][a-zA-Z0-9\-_]{0,61}[a-zA-Z0-9]))*$`) //nolint:lll

func (s Settings) Validate() (err error) {
	err = validate.AllMatchRegex(s.AllowedHosts, hostRegex)
	if err != nil {
		return fmt.Errorf("allowed hosts: %w", err)
	}

	err = validate.AllMatchRegex(s.AddBlockedHosts, hostRegex)
	if err != nil {
		return fmt.Errorf("additional blocked hosts: %w", err)
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
