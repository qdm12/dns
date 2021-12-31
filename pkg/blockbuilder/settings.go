package blockbuilder

import (
	"net/http"
	"strings"

	"github.com/qdm12/gotree"
	"inet.af/netaddr"
)

type Settings struct {
	Client *http.Client
}

func (s *Settings) SetDefaults() {
	if s.Client == nil {
		s.Client = http.DefaultClient
	}
}

type BuildSettings struct {
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

func (s *BuildSettings) SetDefaults() {
	if s.BlockMalicious == nil {
		t := true
		s.BlockMalicious = &t
	}

	if s.BlockAds == nil {
		f := false
		s.BlockAds = &f
	}

	if s.BlockSurveillance == nil {
		f := false
		s.BlockSurveillance = &f
	}
}

func (s *BuildSettings) String() string {
	return s.ToLinesNode().String()
}

func (s *BuildSettings) ToLinesNode() (node *gotree.Node) {
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
