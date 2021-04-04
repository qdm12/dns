package blacklist

import (
	"net"
	"strconv"
	"strings"
)

type BuilderSettings struct {
	BlockMalicious    bool
	BlockAds          bool
	BlockSurveillance bool
	AllowedHosts      []string
	AddBlockedHosts   []string
	AddBlockedIPs     []net.IP
	AddBlockedIPNets  []*net.IPNet
}

func (s *BuilderSettings) String() string {
	const (
		subSection = " |--"
		indent     = "    " // used if lines already contain the subSection
	)
	return strings.Join(s.Lines(indent, subSection), "\n")
}

func (s *BuilderSettings) Lines(indent, subSection string) (lines []string) {
	var blockedCategories []string
	if s.BlockMalicious {
		blockedCategories = append(blockedCategories, "malicious")
	}
	if s.BlockSurveillance {
		blockedCategories = append(blockedCategories, "surveillance")
	}
	if s.BlockAds {
		blockedCategories = append(blockedCategories, "ads")
	}
	lines = append(lines, subSection+"Blocked categories: "+strings.Join(blockedCategories, ", "))

	if len(s.AllowedHosts) > 0 {
		lines = append(lines, subSection+"Additional hostnames blocked: "+
			strconv.Itoa(len(s.AllowedHosts)))
	}

	if len(s.AddBlockedIPs) > 0 {
		lines = append(lines, subSection+"Additional IP addresses blocked: "+
			strconv.Itoa(len(s.AddBlockedIPs)))
	}

	if len(s.AddBlockedIPNets) > 0 {
		lines = append(lines, subSection+"Additional IP networks blocked: "+
			strconv.Itoa(len(s.AddBlockedIPNets)))
	}

	return lines
}
