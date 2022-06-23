package settings

import (
	"fmt"
	"strings"

	"github.com/qdm12/dns/v2/internal/config/defaults"
	"github.com/qdm12/gotree"
	"inet.af/netaddr"
)

type Block struct {
	BlockMalicious       *bool
	BlockAds             *bool
	BlockSurveillance    *bool
	RebindingProtection  *bool
	AllowedHosts         []string
	AllowedIPs           []netaddr.IP
	AllowedIPPrefixes    []netaddr.IPPrefix
	AddBlockedHosts      []string
	AddBlockedIPs        []netaddr.IP
	AddBlockedIPPrefixes []netaddr.IPPrefix
}

func (b *Block) setDefaults() {
	b.BlockMalicious = defaults.BoolPtr(b.BlockMalicious, true)
	b.BlockAds = defaults.BoolPtr(b.BlockAds, false)
	b.BlockSurveillance = defaults.BoolPtr(b.BlockSurveillance, false)
	b.RebindingProtection = defaults.BoolPtr(b.RebindingProtection, true)
}

func (b *Block) validate() (err error) {
	err = checkHostnames(b.AllowedHosts)
	if err != nil {
		return fmt.Errorf("allowed hosts: %w", err)
	}

	err = checkHostnames(b.AddBlockedHosts)
	if err != nil {
		return fmt.Errorf("additional blocked hosts: %w", err)
	}

	return nil
}

func (b *Block) String() string {
	return b.ToLinesNode().String()
}

func (b *Block) ToLinesNode() (node *gotree.Node) { //nolint:cyclop
	node = gotree.New("Filtering:")

	var blockedCategories []string
	if *b.BlockMalicious {
		blockedCategories = append(blockedCategories, "malicious")
	}
	if *b.BlockSurveillance {
		blockedCategories = append(blockedCategories, "surveillance")
	}
	if *b.BlockAds {
		blockedCategories = append(blockedCategories, "ads")
	}

	node.Appendf("Blocked categories: %s", strings.Join(blockedCategories, ", "))

	if *b.RebindingProtection {
		node.Appendf("Rebinding protection: enabled")
	}

	if len(b.AllowedHosts) > 0 {
		node.Appendf("Hostnames unblocked: %d", len(b.AllowedHosts))
	}

	if len(b.AllowedIPs) > 0 {
		node.Appendf("IP addresses unblocked: %d", len(b.AllowedIPs))
	}

	if len(b.AllowedIPPrefixes) > 0 {
		node.Appendf("IP networks unblocked: %d", len(b.AllowedIPPrefixes))
	}

	if len(b.AddBlockedHosts) > 0 {
		node.Appendf("Additional hostnames blocked: %d", len(b.AddBlockedHosts))
	}

	if len(b.AddBlockedIPs) > 0 {
		node.Appendf("Additional IP addresses blocked: %d", len(b.AddBlockedIPs))
	}

	if len(b.AddBlockedIPPrefixes) > 0 {
		node.Appendf("Additional IP networks blocked: %d", len(b.AddBlockedIPPrefixes))
	}

	return node
}
