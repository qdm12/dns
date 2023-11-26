package config

import (
	"fmt"
	"net/netip"
	"regexp"
	"strings"

	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/reader"
	"github.com/qdm12/gosettings/validate"
	"github.com/qdm12/gotree"
)

type Block struct {
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

func (b *Block) setDefaults() {
	b.BlockMalicious = gosettings.DefaultPointer(b.BlockMalicious, true)
	b.BlockAds = gosettings.DefaultPointer(b.BlockAds, false)
	b.BlockSurveillance = gosettings.DefaultPointer(b.BlockSurveillance, false)
}

var regexHostname = regexp.MustCompile(`^([a-zA-Z0-9]|[a-zA-Z0-9_][a-zA-Z0-9\-_]{0,61}[a-zA-Z0-9_])(\.([a-zA-Z0-9]|[a-zA-Z0-9_][a-zA-Z0-9\-_]{0,61}[a-zA-Z0-9]))*$`) //nolint:lll

func (b *Block) validate() (err error) {
	err = validate.AllMatchRegex(b.AllowedHosts, regexHostname)
	if err != nil {
		return fmt.Errorf("allowed hosts: %w", err)
	}

	err = validate.AllMatchRegex(b.AddBlockedHosts, regexHostname)
	if err != nil {
		return fmt.Errorf("additional blocked hosts: %w", err)
	}

	return nil
}

func (b *Block) String() string {
	return b.ToLinesNode().String()
}

func (b *Block) ToLinesNode() (node *gotree.Node) {
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

func (b *Block) read(reader *reader.Reader) (err error) {
	b.BlockMalicious, err = reader.BoolPtr("BLOCK_MALICIOUS")
	if err != nil {
		return err
	}

	b.BlockSurveillance, err = reader.BoolPtr("BLOCK_SURVEILLANCE")
	if err != nil {
		return err
	}

	b.BlockAds, err = reader.BoolPtr("BLOCK_ADS")
	if err != nil {
		return err
	}

	b.AllowedHosts = reader.CSV("ALLOWED_HOSTNAMES")
	b.AddBlockedHosts = reader.CSV("BLOCK_HOSTNAMES")

	b.AllowedIPs, err = reader.CSVNetipAddresses("ALLOWED_IPS")
	if err != nil {
		return err
	}
	b.AddBlockedIPs, err = reader.CSVNetipAddresses("BLOCK_IPS")
	if err != nil {
		return err
	}

	b.AllowedIPPrefixes, err = reader.CSVNetipPrefixes("ALLOWED_CIDRS")
	if err != nil {
		return err
	}
	b.AddBlockedIPPrefixes, err = reader.CSVNetipPrefixes("BLOCK_CIDRS")
	if err != nil {
		return err
	}

	return nil
}
